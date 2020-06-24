// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

// +build bdd

package networktest

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tmrand "github.com/tendermint/tendermint/libs/rand"
	"github.com/tidwall/sjson"
)

// Handles running a testnet using docker-compose.
type Testnet struct {
	Keystore *KeyStore
	chainID  string
	genesis  []byte // Holds the unaltered genesis file that is re-created on every restart.
}

const (
	ContainerCount = 4
	WorkingDir     = "./build/"
	EMD            = WorkingDir + "emd"
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

func NewTestnet() Testnet {
	ks, err := NewKeystore()
	if err != nil {
		panic(err)
	}

	chainID := fmt.Sprintf("localnet-%s", tmrand.Str(6))

	return Testnet{
		Keystore: ks,
		chainID:  chainID,
	}
}

func (t *Testnet) Setup() error {
	err := compileBinaries()
	if err != nil {
		return err
	}

	t.makeTestnet()
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

func (t Testnet) restart(genesismodifier func([]byte) []byte) (func() bool, error) {
	err := dockerComposeDown()
	if err != nil {
		return nil, err
	}

	for i := 0; i < ContainerCount; i++ {
		_, err := execCmdAndWait(EMD, "unsafe-reset-all", "--home", fmt.Sprintf("build/node%d", i))
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

	return dockerComposeUp()
}

func (t Testnet) Teardown() error {
	return dockerComposeDown()
}

func (t Testnet) KillValidator(index int) (string, error) {
	return execCmdAndWait(dockerPath, "kill", fmt.Sprintf("emdnode%v", index))
}

func (t Testnet) ResurrectValidator(index int) (string, error) {
	return execCmdAndWait(dockerPath, "start", fmt.Sprintf("emdnode%v", index))
}

func (t Testnet) GetValidatorLogs(index int) (string, error) {
	return execCmdAndWait(dockerPath, "logs", fmt.Sprintf("emdnode%v", index))
}

func (t Testnet) NewEmcli() Emcli {
	return Emcli{
		chainid:  t.chainID,
		node:     "tcp://localhost:26657",
		keystore: t.Keystore,
	}
}

func (t Testnet) ChainID() string {
	return t.chainID
}

func (t Testnet) makeTestnet() error {
	output, err := execCmdAndWait(EMD,
		"testnet",
		t.chainID,
		t.Keystore.Authority.name,
		"-o", WorkingDir,
		"--keyaccounts", t.Keystore.path)

	if err != nil {
		return err
	}

	t.Keystore.addValidatorKeys(output)
	return nil
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

	window := time.Duration(10 * time.Second).Milliseconds()
	bz, _ = sjson.SetBytes(bz, "app_state.slashing.params.signed_blocks_window_duration", fmt.Sprint(window))

	// Reduce jail time to be able to test unjailing
	unjail := time.Duration(5 * time.Second).Milliseconds()
	bz, _ = sjson.SetBytes(bz, "app_state.slashing.params.downtime_jail_duration", fmt.Sprint(unjail))

	// Start inflation before testnet start in order to have some rewards for NGM stakers.
	inflationLastApplied := time.Now().Add(-2 * time.Hour).UTC().Format(time.RFC3339)
	bz, _ = sjson.SetBytes(bz, "app_state.inflation.assets.last_applied", inflationLastApplied)

	// Set genesis time in the future to allow all docker containers to get up and running
	genesisTime := time.Now().Add(10 * time.Second).UTC().Format(time.RFC3339)
	bz, _ = sjson.SetBytes(bz, "genesis_time", genesisTime)

	t.genesis = bz

	writeGenesisFiles(bz)
}

func compileBinaries() error {
	_, err := execCmdAndWait(makePath, "clean", "build-all")
	if err != nil {
		fmt.Println("Compilation step caused error: ", err)
	}
	return err
}

func dockerComposeUp() (func() bool, error) {
	wait, scanner := createOutputScanner("] Committed state", 30*time.Second)
	return wait, execCmdAndRun(dockerComposePath, []string{"up", "--no-color"}, scanner)
}

func dockerComposeDown() error {
	_, err := execCmdAndWait(dockerComposePath, "kill")
	return err
}

func execCmdAndRun(name string, arguments []string, scanner func(string)) error {
	cmd := exec.Command(name, arguments...)
	err := writeoutput(cmd, scanner)
	if err != nil {
		return err
	}

	return cmd.Start()
}

func execCmdAndWait(name string, arguments ...string) (string, error) {
	cmd := exec.Command(name, arguments...)

	// TODO Look into ways of not always setting this.
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "BUILD_TAGS=fast_consensus")

	var output strings.Builder
	captureOutput := func(s string) {
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
