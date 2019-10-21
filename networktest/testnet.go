// +build bdd

package networktest

import (
	"bufio"
	"context"
	"fmt"
	"github.com/tidwall/sjson"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Handles running a testnet using docker-compose.
type Testnet struct {
	ctx      context.Context
	Keystore *KeyStore
}

const (
	ContainerCount = 4
	WorkingDir     = "./build/"
	EMD            = WorkingDir + "emd-local"
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

func NewTestnetWithContext(ctx context.Context) Testnet {
	ks, err := NewKeystore()
	if err != nil {
		panic(err)
	}

	return Testnet{
		ctx:      ctx,
		Keystore: ks,
	}
}

func NewTestnet() Testnet {
	return NewTestnetWithContext(nil)
}

func (t Testnet) Setup() error {
	err := compileBinaries()
	if err != nil {
		return err
	}

	t.makeTestnet()
	t.updateGenesis()

	return nil
}

func (t Testnet) Start() (func() bool, error) {
	if t.ctx != nil {
		go func() {
			<-t.ctx.Done()
			t.Teardown()
		}()
	}

	return dockerComposeUp()
}

func (t Testnet) Restart() (func() bool, error) {
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

	return dockerComposeUp()
}

func (t Testnet) Teardown() error {
	return dockerComposeDown()
}

func (t Testnet) KillValidator(index int) (string, error) {
	return execCmdAndWait(dockerPath, "kill", fmt.Sprintf("emdnode%v", index))
}

func (t Testnet) WaitFor() {} // Wait for an event, e.g. blocks, special output or ...

func (t Testnet) makeTestnet() error {
	output, err := execCmdAndWait(EMD,
		"testnet",
		"localnet",
		t.Keystore.Authority.name,
		"-o", WorkingDir,
		"--keyaccounts", t.Keystore.path)

	if err != nil {
		return err
	}

	t.Keystore.addValidatorKeys(output)
	return nil
}

func (t Testnet) updateGenesis() {
	genesisPaths := make([]string, 0)
	filepath.Walk(WorkingDir, func(path string, fileinfo os.FileInfo, err error) error {
		if fileinfo.Name() == "genesis.json" {
			genesisPaths = append(genesisPaths, path)
		}
		return nil
	})

	if len(genesisPaths) == 0 {
		panic("Unable to locate genesis.json for testnet.")
	}

	bz, err := ioutil.ReadFile(genesisPaths[0])
	if err != nil {
		panic(err)
	}

	// Tighten slashing conditions.
	bz, _ = sjson.SetBytes(bz, "app_state.slashing.params.min_signed_per_window", "0.3")
	window := time.Duration(10 * time.Second).Milliseconds()
	bz, _ = sjson.SetBytes(bz, "app_state.slashing.params.signed_blocks_window_duration", fmt.Sprint(window))

	for _, path := range genesisPaths {
		err = ioutil.WriteFile(path, bz, 0644)
		if err != nil {
			panic(err)
		}
	}
}

func compileBinaries() error {
	_, err := execCmdAndWait(makePath, "clean", "build-all")
	if err != nil {
		fmt.Println("Compilation step caused error: ", err)
	}
	return err
}

func dockerComposeUp() (func() bool, error) {
	wait, scanner := createOutputScanner("] Committed state", 20*time.Second)
	return wait, execCmdAndRun(dockerComposePath, []string{"up"}, scanner)
}

func dockerComposeDown() error {
	_, err := execCmdAndWait(dockerComposePath, "down")
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
