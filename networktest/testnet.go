// +build bdd

package networktest

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

// Handles running a testnet using docker-compose.
type Testnet struct {
	ctx      context.Context
	keystore *KeyStore
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
		keystore: ks,
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
		err := execCmdAndWait(EMD, "unsafe-reset-all", "--home", fmt.Sprintf("build/node%d", i))
		if err != nil {
			return nil, err
		}
	}

	return dockerComposeUp()
}

// TODO Use context?
func (t Testnet) Teardown() error {
	return dockerComposeDown()
}

func (t Testnet) WaitFor() {} // Wait for an event, e.g. blocks, special output or ...

func (t Testnet) makeTestnet() error {
	return execCmdAndWait(EMD,
		"testnet",
		"localnet",
		t.keystore.Authority.name,
		"-o", "build",
		"--keyaccounts", t.keystore.path)
}

func compileBinaries() error {
	err := execCmdAndWait(makePath, "clean", "build-all")
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
	return execCmdAndWait(dockerComposePath, "down")
}

func execCmdAndRun(name string, arguments []string, scanner func(string)) error {
	cmd := exec.Command(name, arguments...)
	err := writeoutput(cmd, scanner)
	if err != nil {
		return err
	}

	return cmd.Start()
}

func execCmdAndWait(name string, arguments ...string) error {
	cmd := exec.Command(name, arguments...)

	err := writeoutput(cmd)
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}

func writeoutput(cmd *exec.Cmd, filters ...func(string)) error {
	stderrReader, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	go func() {
		scanner := bufio.NewScanner(stderrReader)
		for scanner.Scan() {
			s := scanner.Text()
			fmt.Fprintf(output, "stderr | %s\n", s)

			for _, filter := range filters {
				filter(s)
			}
		}
	}()

	stdoutReader, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	go func() {
		scanner := bufio.NewScanner(stdoutReader)
		for scanner.Scan() {
			s := scanner.Text()
			fmt.Fprintf(output, "stdout | %s\n", s)

			for _, filter := range filters {
				filter(s)
			}
		}
	}()

	return nil
}
