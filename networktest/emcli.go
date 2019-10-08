package networktest

import (
	"io"
	"os/exec"
	"strings"
)

const (
	EMCLI = "./build/emcli-local"
)

type Emcli struct {
	node     string
	chainid  string
	keystore *KeyStore
}

func NewEmcli(keystore *KeyStore) Emcli {
	return Emcli{
		chainid:  "localnet",
		node:     "tcp://localhost:26657",
		keystore: keystore,
	}
}

func (cli Emcli) QueryIssuers() ([]byte, error) {
	return execCmdAndCollectResponse(cli.addNetworkFlags("q", "issuers")...)
}

func (cli Emcli) QueryInflation() ([]byte, error) {
	return execCmdAndCollectResponse(cli.addNetworkFlags("q", "inflation")...)
}

func (cli Emcli) AuthorityCreateIssuer(issuerKey string, denoms ...string) ([]byte, error) {
	args := cli.addNetworkFlags("authority", "create-issuer", "authoritykey", issuerKey, strings.Join(denoms, ","), "--yes")
	return execCmdWithInput(args, "pwd12345\n")
}

func execCmdWithInput(arguments []string, input string) ([]byte, error) {
	cmd := exec.Command(EMCLI, arguments...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
	}

	io.WriteString(stdin, input)

	if err != nil {
		panic(err)
	}

	return cmd.CombinedOutput()
}

func execCmdAndCollectResponse(arguments ...string) ([]byte, error) {
	return exec.Command(EMCLI, arguments...).CombinedOutput()
}

func (cli Emcli) addNetworkFlags(arguments ...string) []string {
	return append(arguments,
		"--node", cli.node,
		"--chain-id", cli.chainid,
		"--home", cli.keystore.path,
		"--output", "json",
	)
}
