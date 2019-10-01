// +build bdd

package network_test

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
)

type Testnet struct {
	ctx context.Context
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
	return Testnet{
		ctx: ctx,
	}
}

func NewTestnet() Testnet {
	return Testnet{}
}

func (t Testnet) Setup() error {
	//make clean build-linux && ./emd.sh testnet localnet faucet -o build  --keyaccounts ~/.emcli/
	fns := []func() error{
		compileBinaries,
		makeTestnet,
	}

	for _, fn := range fns {
		err := fn()
		if err != nil {
			return err
		}
	}

	return nil
}

func (t Testnet) Start() error {
	if t.ctx != nil {
		go func() {
			<-t.ctx.Done()
			t.Teardown()
		}()
	}

	return dockerComposeUp()
}

func (t Testnet) Restart() error {
	err := dockerComposeDown()
	if err != nil {
		return err
	}

	for i := 0; i < ContainerCount; i++ {
		err := execCmdAndWait(EMD, "unsafe-reset-all", "--home", fmt.Sprintf("build/node%d", i))
		if err != nil {
			return err
		}
	}

	return dockerComposeUp()
}

// TODO Use context?
func (t Testnet) Teardown() error {
	return dockerComposeDown()
}

func (t Testnet) WaitFor() {} // Wait for an event, e.g. blocks, special output or ...

func makeTestnet() error {
	return execCmdAndWait(EMD, "testnet", "localnet", "master", "-o", "build", "--keyaccounts", "./network_test/testdata/")
}

func compileBinaries() error {
	return execCmdAndWait(makePath, "clean", "build-all")
}

func dockerComposeUp() error {
	return execCmdAndRun(dockerComposePath, "up")
}

func dockerComposeDown() error {
	return execCmdAndWait(dockerComposePath, "down")
}

func execCmdAndRun(name string, arguments ...string) error {
	cmd := exec.Command(name, arguments...)
	err := writeoutput(cmd)
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

func writeoutput(cmd *exec.Cmd) error {
	stderrReader, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	go func() {
		scanner := bufio.NewScanner(stderrReader)
		for scanner.Scan() {
			s := scanner.Text()
			fmt.Fprintf(output, "stderr | %s\n", s)
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
		}
	}()

	return nil
}
