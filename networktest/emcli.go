package networktest

import (
	"os/exec"
	"strings"

	"github.com/cosmos/cosmos-sdk/crypto/keys"
)

const (
	EMCLI = "./build/emcli-local"
)

type Emcli struct {
	node    string
	chainid string
	homedir string
}

type Key struct {
	name    string
	keybase keys.Keybase
}

func NewEmcli() Emcli {
	emcli := Emcli{}

	emcli.chainid = "localnet"
	emcli.node = "tcp://localhost:26657"
	emcli.homedir = "./networktest/testdata/"

	return emcli
}

func (cli Emcli) QueryIssuers() ([]byte, error) {
	return execCmdAndCollectResponse(cli.addNetworkFlags("q", "issuers")...)
}

func (cli Emcli) QueryInflation() ([]byte, error) {
	return execCmdAndCollectResponse(cli.addNetworkFlags("q", "inflation")...)
}

func (cli Emcli) AuthorityCreateIssuer(issuerKey string, denoms ...string) {
	execCmdAndCollectResponse("authority", "create-issuer", "master", issuerKey, strings.Join(denoms, ","))
}

func execCmdAndCollectResponse(arguments ...string) ([]byte, error) {
	//fmt.Printf("Estimated command: %v %v\n", EMCLI, strings.Join(arguments, " "))
	return exec.Command(EMCLI, arguments...).CombinedOutput()
}

func (cli Emcli) addNetworkFlags(arguments ...string) []string {
	return append(arguments,
		"--node", cli.node,
		"--chain-id", cli.chainid,
		"--home", cli.homedir,
		"--output", "json",
	)
}
