package networktest

import (
	"fmt"
	"github.com/tidwall/gjson"
	"io"
	"os/exec"
	"strconv"
	"strings"
)

const (
	EMCLI = "./build/emcli-local"

	// gjson paths
	QGetCreditEUR  = "value.Credit.#(denom==\"x2eur\").amount"
	QGetBalanceEUR = "value.Account.value.coins.#(denom==\"x2eur\").amount"
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
	return execCmdAndCollectResponse(cli.addQueryFlags("q", "issuers"))
}

func (cli Emcli) QueryInflation() ([]byte, error) {
	return execCmdAndCollectResponse(cli.addQueryFlags("q", "inflation"))
}

func (cli Emcli) AuthorityCreateIssuer(authority, issuer Key, denoms ...string) (string, error) {
	args := cli.addTransactionFlags("authority", "create-issuer", authority.name, issuer.GetAddress(), strings.Join(denoms, ","))
	return execCmdWithInput(args, KeyPwd)
}

func (cli Emcli) QueryTransaction(txhash string) ([]byte, error) {
	args := cli.addQueryFlags("query", "tx", txhash)
	return execCmdAndCollectResponse(args)
}

// NOTE Hardcoded to x2eur for now.
func (cli Emcli) QueryAccount(account string) (balance, credit int, err error) {
	args := cli.addQueryFlags("query", "account", account)
	bz, err := execCmdAndCollectResponse(args)
	if err != nil {
		return 0, 0, err
	}

	queryresponse := gjson.ParseBytes(bz)

	v := queryresponse.Get(QGetBalanceEUR)
	balance, _ = strconv.Atoi(v.Str)

	v = queryresponse.Get(QGetCreditEUR)
	if v.Exists() {
		credit, _ = strconv.Atoi(v.Str)
	}

	return
}

func (cli Emcli) QueryAccountJson(account string) ([]byte, error) {
	args := cli.addQueryFlags("query", "account", account)
	return execCmdAndCollectResponse(args)
}

func (cli Emcli) IssuerIncreaseCredit(issuer, liquidityprovider Key, amount string) (string, error) {
	args := cli.addTransactionFlags("issuer", "increase-credit", issuer.name, liquidityprovider.GetAddress(), amount)
	return execCmdWithInput(args, KeyPwd)
}

func (cli Emcli) IssuerRevokeCredit(issuer, liquidityprovider Key) (string, error) {
	args := cli.addTransactionFlags("issuer", "revoke-credit", issuer.name, liquidityprovider.GetAddress())
	return execCmdWithInput(args, KeyPwd)
}

func (cli Emcli) IssuerDecreaseCredit(issuer, liquidityprovider Key, amount string) (string, error) {
	args := cli.addTransactionFlags("issuer", "decrease-credit", issuer.name, liquidityprovider.GetAddress(), amount)
	return execCmdWithInput(args, KeyPwd)
}

func (cli Emcli) IssuerSetInflation(issuer Key, denom string, inflation string) (string, error) {
	args := cli.addTransactionFlags("issuer", "set-inflation", issuer.name, denom, inflation)
	return execCmdWithInput(args, KeyPwd)
}

func (cli Emcli) LiquidityProviderMint(key Key, amount string) (string, error) {
	args := cli.addTransactionFlags("liquidityprovider", "mint", amount, "--from", key.name)
	return execCmdWithInput(args, KeyPwd)
}

func (cli Emcli) LiquidityProviderBurn(key Key, amount string) (string, error) {
	args := cli.addTransactionFlags("liquidityprovider", "burn", amount, "--from", key.name)
	return execCmdWithInput(args, KeyPwd)
}

func extractTxHash(bz []byte) (string, error) {
	txhash := gjson.ParseBytes(bz).Get("txhash")
	if txhash.Exists() {
		return txhash.Str, nil
	}

	return "", fmt.Errorf("could not find txhash in response %v", string(bz))
}

func execCmdWithInput(arguments []string, input string) (string, error) {
	//fmt.Println(" *** Running command: ", EMCLI, strings.Join(arguments, " "))
	cmd := exec.Command(EMCLI, arguments...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}

	_, err = io.WriteString(stdin, input+"\n")
	if err != nil {
		return "", err
	}

	bz, err := cmd.CombinedOutput()
	if err != nil {
		//fmt.Println(" *** Running command: ", EMCLI, strings.Join(arguments, " "))
		//fmt.Println(" *** Output", string(bz))
		return "", err
	}

	return extractTxHash(bz)
}

func execCmdAndCollectResponse(arguments []string) ([]byte, error) {
	//fmt.Println(" *** Running command: ", EMCLI, strings.Join(arguments, " "))
	bz, err := exec.Command(EMCLI, arguments...).CombinedOutput()
	//fmt.Println(" *** Output: ", string(bz))
	return bz, err
}

func (cli Emcli) addQueryFlags(arguments ...string) []string {
	return cli.addNetworkFlags(arguments)
}

func (cli Emcli) addTransactionFlags(arguments ...string) []string {
	arguments = append(arguments,
		"--broadcast-mode", "block",
		"--home", cli.keystore.path,
		"--yes",
	)

	return cli.addNetworkFlags(arguments)
}

func (cli Emcli) addNetworkFlags(arguments []string) []string {
	return append(arguments,
		"--node", cli.node,
		"--chain-id", cli.chainid,
		"--output", "json",
	)
}
