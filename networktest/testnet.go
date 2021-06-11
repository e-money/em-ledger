// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

// +build bdd

package networktest

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	tmrand "github.com/tendermint/tendermint/libs/rand"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// Handles running a testnet using docker-compose.
type Testnet struct {
	Keystore *KeyStore
	chainID  string
	genesis  []byte // Holds the unaltered genesis file that is re-created on every restart.
	sync.Mutex
}

const (
	ContainerCount = 4
	WorkingDir     = "./build/"
	EMD            = WorkingDir + "emd"
	DefaultNode    = "tcp://localhost:26657"
)

var (
	dockerComposePath string
	dockerPath        string
	makePath          string
	output            io.Writer = os.Stdout // Override to make tests quiet
)

func init() {
	dockerComposePath = locateExecutable("docker-compose")
	dockerPath = locateExecutable("docker")
	makePath = locateExecutable("make")
}

func locateExecutable(name string) (path string) {
	path, err := exec.LookPath(name)
	if path == "" {
		fmt.Printf("Unable to locate %s\n", name)
	}
	if err != nil {
		panic(err)
	}
	return
}

// NewTestnet launches a new testnet with random or fixed location and chain id
// or syncs with an existing one with chain ID: LocalNetReuse
func NewTestnet() Testnet {
	reusableIsUp := reusableNetUp()
	launchReusable := os.Getenv(startForReUseEnv) == "1"
	ks, err := NewKeystore(launchReusable, reusableIsUp)
	if err != nil {
		panic(err)
	}

	chainID := LocalNetReuse
	if !launchReusable && !reusableIsUp {
		chainID = fmt.Sprintf("localnet-%s", tmrand.Str(6))
	}

	return Testnet{
		Keystore: ks,
		chainID:  chainID,
	}
}

func (t *Testnet) Setup() error {
	if reusableNetUp() {
		return nil
	}

	err := t.compileBinaries()
	if err != nil {
		return err
	}

	err = t.makeTestnet()
	if err != nil {
		return err
	}

	t.updateGenesis()

	return nil
}

func writeGenesisFiles(newGenesisFile []byte) error {
	return filepath.Walk(WorkingDir, func(path string, fileinfo os.FileInfo, err error) error {
		if fileinfo.Name() == "genesis.json" {
			err := ioutil.WriteFile(path, newGenesisFile, 0644)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (t Testnet) Restart() (func() bool, error) {
	return t.restart(nil)
}

func (t Testnet) RestartWithModifications(genesisModifier func([]byte) []byte) (func() bool, error) {
	return t.restart(genesisModifier)
}

func (t *Testnet) restart(genesismodifier func([]byte) []byte) (func() bool, error) {
	if reusableNetUp() {
		return func() bool { return true }, nil
	}
	err := t.dockerComposeDown()
	if err != nil {
		return nil, err
	}

	for i := 0; i < ContainerCount; i++ {
		_, err := t.execCmdAndWait(EMD, "unsafe-reset-all", "--home", fmt.Sprintf("build/node%d", i))
		if err != nil {
			return nil, err
		}
	}

	if genesismodifier != nil {
		modifiedGenesis := make([]byte, len(t.genesis))
		copy(modifiedGenesis, t.genesis)
		modifiedGenesis = genesismodifier(modifiedGenesis)
		writeGenesisFiles(modifiedGenesis)
	} else {
		// Restore the default genesis files
		writeGenesisFiles(t.genesis)
	}

	return t.dockerComposeUp()
}

func (t *Testnet) Teardown() error {
	if reusableNetUp() {
		return nil
	}
	return t.dockerComposeDown()
}

func (t *Testnet) KillValidator(index int) (string, error) {
	return t.execCmdAndWait(dockerComposePath, "kill", fmt.Sprintf("emdnode%v", index))
}

func (t *Testnet) ResurrectValidator(index int) (string, error) {
	return t.execCmdAndWait(dockerComposePath, "start", fmt.Sprintf("emdnode%v", index))
}

func (t *Testnet) GetValidatorLogs(index int) (string, error) {
	return t.execCmdAndWait(dockerComposePath, "logs", fmt.Sprintf("emdnode%v", index))
}

func (t *Testnet) NewEmcli() Emcli {
	return Emcli{
		chainid:  t.chainID,
		node:     DefaultNode,
		keystore: t.Keystore,
	}
}

func (t *Testnet) ChainID() string {
	return t.chainID
}

func (t *Testnet) makeTestnet() error {
	numNodes := 4

	if reusableNetUp() {
		return nil
	}

	_, err := t.execCmdAndWait(
		EMD,
		"testnet",
		t.Keystore.Authority.name,
		"--chain-id", t.chainID,
		"-o", WorkingDir,
		"--keyring-backend", "test",
		"--starting-ip-address", "192.168.10.2",
		"--keyaccounts", t.Keystore.path,
		"--commit-timeout", "1500ms",
		"--v", strconv.Itoa(numNodes),
		"--minimum-gas-prices", "")

	if err != nil {
		return err
	}

	t.Keystore.addValidatorKeys(WorkingDir, numNodes)
	return nil
}

func reusableNetUp() bool {
	status, err := exec.Command(EMCLI, "status").CombinedOutput()
	if err != nil {
		return false
	}

	if strings.Contains(string(status), LocalNetReuse) {
		return true
	}

	return false
}

func (t *Testnet) updateGenesis() {
	var genesisPath string
	filepath.Walk(WorkingDir, func(path string, fileinfo os.FileInfo, err error) error {
		if genesisPath != "" {
			return filepath.SkipDir
		}

		if fileinfo.Name() == "genesis.json" {
			genesisPath = path
		}
		return nil
	})

	if genesisPath == "" {
		panic("Unable to locate genesis.json for testnet.")
	}

	bz, err := ioutil.ReadFile(genesisPath)
	if err != nil {
		panic(err)
	}

	// Tighten slashing conditions.
	bz, _ = sjson.SetBytes(bz, "app_state.slashing.params.min_signed_per_window", "0.3")

	bz, _ = sjson.SetBytes(bz, "app_state.slashing.params.signed_blocks_window", (10 * time.Second).Nanoseconds())

	// Reduce jail time to be able to test unjailing
	bz, _ = sjson.SetBytes(bz, "app_state.slashing.params.downtime_jail_duration", "5s")

	// Start inflation before testnet start in order to have some rewards for NGM stakers.
	inflationLastApplied := time.Now().Add(-2 * time.Hour).UTC().Format(time.RFC3339)
	bz, _ = sjson.SetBytes(bz, "app_state.inflation.assets.last_applied", inflationLastApplied)

	// Set genesis time in the future to allow all docker containers to get up and running
	genesisTime := time.Now().Add(10 * time.Second).UTC().Format(time.RFC3339)
	bz, _ = sjson.SetBytes(bz, "genesis_time", genesisTime)

	t.genesis = bz

	writeGenesisFiles(bz)
}

// IncChain waits till the chain increments by the requested blocks or times out.
func IncChain(delta int64) (int64, error) {
	height, err := GetHeight()
	if err != nil {
		return height, err
	}

	return IncChainWithExpiration(
		height+delta,
		// generous and unlikely to exhaust
		time.Duration(delta*5)*time.Second,
	)
}

func IncChainWithExpiration(height int64, sleepDur time.Duration) (int64, error) {
	newHeight, err := WaitForHeightWithTimeout(
		height+1, sleepDur,
	)

	return newHeight, err
}

func ChainBlockHash() (string, error) {
	status, err := chainStatus()
	if err != nil {
		return "", err
	}

	blockHash := gjson.ParseBytes(status).Get("SyncInfo.latest_block_hash").Str

	return blockHash, nil
}

func GetHeight() (int64, error) {
	status, err := chainStatus()
	if err != nil {
		return -1, err
	}

	height := gjson.ParseBytes(status).Get("SyncInfo.latest_block_height").Int()

	return height, nil
}

func chainStatus() ([]byte, error) {
	return exec.Command(EMCLI, "status").CombinedOutput()
}

// WaitForHeightWithTimeout waits till the chain reaches the requested height
// or times out whichever occurs first.
func WaitForHeightWithTimeout(requestedHeight int64, t time.Duration) (int64, error) {
	// Half a sec + emd invoke + result processing
	ticker := time.NewTicker(500 * time.Millisecond)
	timeout := time.After(t)

	var blockHeight int64 = -1

	for {
		select {
		case <-timeout:
			ticker.Stop()
			return blockHeight, fmt.Errorf(
				"timeout at height %d, before reaching height:%d",
				blockHeight,
				requestedHeight)
		case <-ticker.C:
			height, err := GetHeight()
			if err != nil {
				fmt.Print(".")
				continue
			}
			if height >= requestedHeight {
				return height, nil
			}
		}
	}
}

// WaitForListenerWithTimeout waits till the chain is able to return an event
// listener or times out.
func WaitForListenerWithTimeout(t time.Duration) (listener EventListener, err error) {
	ticker := time.NewTicker(500 * time.Millisecond)
	timeout := time.After(t)

	for {
		select {
		case <-timeout:
			ticker.Stop()
			return EventListener{},
			errors.New("timeout before receiving an event listener")
		case <-ticker.C:
			listener, err = NewEventListener()
			if err != nil {
				fmt.Print(".")
				continue
			}

			newBlocker, err := listener.AwaitNewBlock()
			if err != nil {
				fmt.Print(".")
				continue
			}
			if !newBlocker() {
				fmt.Print(".")
				continue
			}

			return listener, nil
		}
	}
}

func (t *Testnet)compileBinaries() error {
	_, err := t.execCmdAndWait(makePath, "clean", "build-all")
	if err != nil {
		fmt.Println("Compilation step caused error: ", err)
	}
	return err
}

func (t *Testnet)dockerComposeUp() (func() bool, error) {
	wait := func() bool {
		_, err := WaitForHeightWithTimeout(1, 30*time.Second)
		return err == nil
	}
	_, scanner := createOutputScanner("committed state", 30*time.Second)
	return wait, t.execCmdAndRun(dockerComposePath, []string{"up"}, scanner)
}

func (t *Testnet)dockerComposeDown() error {
	_, err := t.execCmdAndWait(dockerComposePath, "kill")
	return err
}

func (t *Testnet)execCmdAndRun(name string, arguments []string, scanner func(string)) error {
	cmd := exec.Command(name, arguments...)
	err := writeoutput(cmd, scanner)
	if err != nil {
		return err
	}

	return cmd.Start()
}

func (t *Testnet)execCmdAndWait(name string, arguments ...string) (string, error) {
	cmd := exec.Command(name, arguments...)

	// TODO Look into ways of not always setting this.
	// todo (reviewer): dropped "fast_consensus" build tag

	var output strings.Builder
	captureOutput := func(s string) {
		t.Lock()
		defer t.Unlock()
		output.WriteString(s)
		output.WriteRune('\n')
	}

	err := writeoutput(cmd, captureOutput)
	if err != nil {
		return "", err
	}

	err = cmd.Start()
	if err != nil {
		return "", err
	}

	err = cmd.Wait()
	if err != nil {
		return "", err
	}

	return output.String(), nil
}

func writeoutput(cmd *exec.Cmd, filters ...func(string)) error {
	stderrReader, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	stdoutReader, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	go createOutputHandler("stderr", stderrReader, filters)
	go createOutputHandler("stdout", stdoutReader, filters)
	return nil
}

func createOutputHandler(prefix string, reader io.Reader, filters []func(string)) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		s := scanner.Text()
		fmt.Fprintf(output, "%s | %s\n", prefix, s)

		for _, filter := range filters {
			filter(s)
		}
	}
}
