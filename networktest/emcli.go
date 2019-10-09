package networktest

import (
	"fmt"
	"github.com/tidwall/gjson"
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
	return execCmdAndCollectResponse(cli.addNetworkFlags("q", "issuers"))
}

func (cli Emcli) QueryInflation() ([]byte, error) {
	return execCmdAndCollectResponse(cli.addNetworkFlags("q", "inflation"))
}

func (cli Emcli) AuthorityCreateIssuer(issuerKey Key, denoms ...string) ([]byte, error) {
	args := cli.addNetworkFlags("authority", "create-issuer", "authoritykey", issuerKey.GetAddress(), strings.Join(denoms, ","), "--yes")
	return execCmdWithInput(args, "pwd12345\n")
}

func (cli Emcli) QueryTransaction(txhash string) ([]byte, error) {
	args := cli.addNetworkFlags("query", "tx", txhash)
	return execCmdAndCollectResponse(args)
}

func (cli Emcli) QueryAccount(account string) ([]byte, error) {
	args := cli.addNetworkFlags("query", "account", account)
	return execCmdAndCollectResponse(args)
}

func (cli Emcli) IssuerIncreaseCredit(issuer, liquidityprovider Key, amount string) (string, error) {
	args := cli.addNetworkFlags("issuer", "increase-credit", issuer.name, liquidityprovider.GetAddress(), amount, "--yes")
	bz, err := execCmdWithInput(args, "pwd12345\n")
	if err != nil {
		return "", err
	}

	return extractTxHash(bz)
}

func (cli Emcli) IssuerDecreaseCredit(issuer, liquidityprovider Key, amount string) ([]byte, error) {
	args := cli.addNetworkFlags("issuer", "decrease-credit", issuer.name, liquidityprovider.GetAddress(), amount, "--yes")
	return execCmdWithInput(args, "pwd12345\n")
}

func (cli Emcli) LiquidityProviderMint(key Key, amount string) (string, error) {
	args := cli.addNetworkFlags("liquidityprovider", "mint", amount, "--from", key.name, "--yes")
	bz, err := execCmdWithInput(args, "pwd12345\n")
	if err != nil {
		return "", err
	}

	return extractTxHash(bz)
}

func (cli Emcli) QueryTransactionSucessful(txhash string) (bool, error) {
	bz, err := cli.QueryTransaction(txhash)
	if err != nil {
		return false, err
	}

	return gjson.ParseBytes(bz).Get("logs.0.success").Bool(), nil
}

func extractTxHash(bz []byte) (string, error) {
	txhash := gjson.ParseBytes(bz).Get("txhash")
	if txhash.Exists() {
		return txhash.Str, nil
	}

	return "", fmt.Errorf("could not find txhash in response %v", string(bz))
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

func execCmdAndCollectResponse(arguments []string) ([]byte, error) {
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
