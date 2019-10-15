// +build bdd

package networktest

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
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
	EMD            = "./build/emd-local"
)

var (
	dockerComposePath string
	makePath          string
	output            io.Writer = os.Stdout // Override to make tests quiet
)

func init() {
	var err error

	dockerComposePath, err = exec.LookPath("docker-compose")
	if dockerComposePath == "" {
		fmt.Println("Unable to locate docker-compose")
	}
	if err != nil {
		panic(err)
	}

	makePath, err = exec.LookPath("make")
	if makePath == "" {
		fmt.Println("Unable to locate make")
	}
	if err != nil {
		panic(err)
	}
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

func (t Testnet) WaitFor() {} // Wait for an event, e.g. blocks, special output or ...

func (t Testnet) makeTestnet() error {
	output, err := execCmdAndWait(EMD,
		"testnet",
		"localnet",
		t.Keystore.Authority.name,
		"-o", "build",
		"--keyaccounts", t.Keystore.path)

	if err != nil {
		return err
	}

	t.Keystore.addValidatorKeys(output)
	return nil
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
