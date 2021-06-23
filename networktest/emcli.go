// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package networktest

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tidwall/gjson"
)

const (
	// todo (reviewer) : emcli was merged into emd:
	EMCLI = "./build/emd"

	// gjson paths
	QGetMintableEUR = "mintable.#(denom==\"eeur\").amount"
	QGetBalanceEUR  = "balances.#(denom==\"eeur\").amount"
)

type Emcli struct {
	node     string
	chainid  string
	keystore *KeyStore
}

func (cli Emcli) QueryIssuers() ([]byte, error) {
	return execCmdAndCollectResponse(cli.addQueryFlags("q", "issuers"))
}

func (cli Emcli) QueryInflation() ([]byte, error) {
	return execCmdAndCollectResponse(cli.addQueryFlags("q", "inflation"))
}

func (cli Emcli) Send(from, to Key, amount string) (string, bool, error) {
	args := cli.addTransactionFlags("tx", "bank", "send", from.name, to.GetAddress(), amount)
	return execCmdWithInput(args, KeyPwd)
}

func (cli Emcli) AuthorityCreateIssuer(authority, issuer Key, denoms ...string) (string, bool, error) {
	args := cli.addTransactionFlags("tx", "authority", "create-issuer", authority.name, issuer.GetAddress(), strings.Join(denoms, ","))
	return execCmdWithInput(args, KeyPwd)
}

func (cli Emcli) AuthorityDestroyIssuer(authority, issuer Key) (string, bool, error) {
	args := cli.addTransactionFlags("tx", "authority", "destroy-issuer", authority.name, issuer.GetAddress())
	return execCmdWithInput(args, KeyPwd)
}

func (cli Emcli) CustomCommand(params ...string) (string, error) {
	args := cli.addTransactionFlags(params...)
	return execCmdCollectOutput(args, KeyPwd)
}

func (cli Emcli) AuthoritySetMinGasPricesMulti(from, minGasPrices string, params ...string) (string, error) {
	args := cli.addTransactionFlags("tx", "authority", "set-gas-prices", from, minGasPrices)
	args = append(args, params...)
	return execCmdCollectOutput(args, KeyPwd)
}

func (cli Emcli) AuthoritySetMinGasPrices(authority Key, minGasPrices string, params ...string) (string, bool, error) {
	args := cli.addTransactionFlags("tx", "authority", "set-gas-prices", authority.name, minGasPrices)
	args = append(args, params...)
	return execCmdWithInput(args, KeyPwd)
}

func (cli Emcli) QueryBuybackBalance() ([]byte, error) {
	args := cli.addQueryFlags("query", "buyback", "balance")
	return execCmdAndCollectResponse(args)
}

func (cli Emcli) QueryMinGasPrices() ([]byte, error) {
	args := cli.addQueryFlags("query", "authority", "gas-prices")
	return execCmdAndCollectResponse(args)
}

func (cli Emcli) QueryTransaction(txhash string) ([]byte, error) {
	args := cli.addQueryFlags("query", "tx", txhash)
	return execCmdAndCollectResponse(args)
}

func (cli Emcli) QueryValidatorCommission(validator string) ([]byte, error) {
	args := cli.addQueryFlags("query", "distribution", "commission", validator)
	return execCmdAndCollectResponse(args)
}

func (cli Emcli) QueryRewards(delegator string) (gjson.Result, error) {
	args := cli.addQueryFlags("query", "distribution", "rewards", delegator)

	bz, err := execCmdAndCollectResponse(args)
	if err != nil {
		return gjson.Result{}, err
	}

	return gjson.ParseBytes(bz), nil
}

// NOTE Hardcoded to eeur for now.
func (cli Emcli) QueryBalance(account string) (balance int, err error) {
	args := cli.addQueryFlags("query", "bank", "balances", account)
	bz, err := execCmdAndCollectResponse(args)
	if err != nil {
		return 0, err
	}

	queryresponse := gjson.ParseBytes(bz)

	v := queryresponse.Get(QGetBalanceEUR)
	if v.Exists() {
		balance, _ = strconv.Atoi(v.Str)
	}

	return
}

// QueryBalanceDenom retrieve Balance by Denom
func (cli Emcli) QueryBalanceDenom(account, denom string) (balance int, err error) {
	args := cli.addQueryFlags("query", "bank", "balances", account)
	bz, err := execCmdAndCollectResponse(args)
	if err != nil {
		return 0, err
	}

	queryresponse := gjson.ParseBytes(bz)

	denomQ := fmt.Sprintf(`balances.#(denom=="%s").amount`, denom)
	v := queryresponse.Get(denomQ)
	if v.Exists() {
		balance, _ = strconv.Atoi(v.Str)
	}

	return
}

func (cli Emcli) QueryAccount(account string) (mintable int, err error) {
	args := cli.addQueryFlags("query", "account", account)
	_, err = execCmdAndCollectResponse(args)
	if err != nil {
		return 0, err
	}

	return
}

// NOTE Hardcoded to eeur for now.
func (cli Emcli) QueryMintable(account string) (mintable int, err error) {
	args := cli.addQueryFlags("query", "liquidityprovider", "mintable", account)
	bz, err := execCmdAndCollectResponse(args)
	if err != nil {
		return 0, err
	}

	queryresponse := gjson.ParseBytes(bz)

	v := queryresponse.Get(QGetMintableEUR)
	if v.Exists() {
		mintable, _ = strconv.Atoi(v.Str)
	}

	return
}

func (cli Emcli) QueryTotalSupply() ([]byte, error) {
	args := cli.addQueryFlags("query", "bank", "total")
	return execCmdAndCollectResponse(args)
}

func (cli Emcli) QueryAccountJson(account string) ([]byte, error) {
	args := cli.addQueryFlags("query", "account", account)
	return execCmdAndCollectResponse(args)
}

func (cli Emcli) QueryMintableJson(account string) ([]byte, error) {
	args := cli.addQueryFlags("query", "liquidityprovider", "mintable", account)
	return execCmdAndCollectResponse(args)
}

func (cli Emcli) QueryMarketInstruments() ([]byte, error) {
	args := cli.addQueryFlags("query", "market", "instruments")
	return execCmdAndCollectResponse(args)
}

func (cli Emcli) QueryMarketInstrument(source, destination string) ([]byte, error) {
	args := cli.addQueryFlags("query", "market", "instrument", source, destination)
	return execCmdAndCollectResponse(args)
}

func (cli Emcli) QueryMarketByAccount(account string) ([]byte, error) {
	args := cli.addQueryFlags("query", "market", "account", account)
	return execCmdAndCollectResponse(args)
}

func (cli Emcli) QueryDelegationsTo(validator string) ([]byte, error) {
	args := cli.addQueryFlags("query", "staking", "delegations-to", validator)
	return execCmdAndCollectResponse(args)
}

func (cli Emcli) QueryValidators() (gjson.Result, error) {
	args := cli.addQueryFlags("query", "staking", "validators")
	bz, err := execCmdAndCollectResponse(args)
	if err != nil {
		return gjson.Result{}, err
	}

	return gjson.ParseBytes(bz), nil
}

func (cli Emcli) BEP3ListSwaps() (string, error) {
	args := cli.addQueryFlags("query", "bep3", "swaps")
	bz, err := execCmdAndCollectResponse(args)

	return string(bz), err
}

func (cli Emcli) BEP3SupplyOf(denom string) (string, error) {
	args := cli.addQueryFlags("query", "bep3", "supply", denom)
	bz, err := execCmdAndCollectResponse(args)

	return string(bz), err
}

func (cli Emcli) QueryDelegations(account string) ([]byte, error) {
	args := cli.addQueryFlags("query", "staking", "delegations", account)
	return execCmdAndCollectResponse(args)
}

func (cli Emcli) SignTranscation(txPath, fromAddress, multisigAddress string) (string, error) {
	args := cli.addTransactionFlags("tx", "sign", txPath, "--from", fromAddress, "--multisig", multisigAddress)
	return execCmdCollectOutput(args, KeyPwd)
}

func (cli Emcli) IssuerIncreaseMintableAmount(issuer, liquidityprovider Key, amount string) (string, bool, error) {
	args := cli.addTransactionFlags("tx", "issuer", "increase-mintable", issuer.name, liquidityprovider.GetAddress(), amount)
	return execCmdWithInput(args, KeyPwd)
}

func (cli Emcli) IssuerRevokeMinting(issuer, liquidityprovider Key) (string, bool, error) {
	args := cli.addTransactionFlags("tx", "issuer", "revoke-mint", issuer.name, liquidityprovider.GetAddress())
	return execCmdWithInput(args, KeyPwd)
}

func (cli Emcli) IssuerDecreaseMintableAmount(issuer, liquidityprovider Key, amount string) (string, bool, error) {
	args := cli.addTransactionFlags("tx", "issuer", "decrease-mintable", issuer.name, liquidityprovider.GetAddress(), amount)
	return execCmdWithInput(args, KeyPwd)
}

func (cli Emcli) IssuerSetInflation(issuer Key, denom string, inflation string) (string, bool, error) {
	args := cli.addTransactionFlags("tx", "issuer", "set-inflation", issuer.name, denom, inflation)
	return execCmdWithInput(args, KeyPwd)
}

func (cli Emcli) LiquidityProviderMint(key Key, amount string) (string, bool, error) {
	args := cli.addTransactionFlags("tx", "liquidityprovider", "mint", key.name, amount)
	return execCmdWithInput(args, KeyPwd)
}

func (cli Emcli) LiquidityProviderBurn(key Key, amount string) (string, bool, error) {
	args := cli.addTransactionFlags("tx", "liquidityprovider", "burn", key.name, amount)
	return execCmdWithInput(args, KeyPwd)
}

func (cli Emcli) MarketAddLimitOrder(key Key, source, destination, cid string, moreflags ...string) (string, bool, error) {
	args := cli.addTransactionFlags("tx", "market", "add-limit", source, destination, cid, "--from", key.name)
	args = append(args, moreflags...)
	return execCmdWithInput(args, KeyPwd)
}

func (cli Emcli) MarketAddMarketOrder(key Key, sourceDenom, destination, cid string, slippage sdk.Dec, moreflags ...string) (string, bool, error) {
	args := cli.addTransactionFlags("tx", "market", "add-market", sourceDenom, destination, slippage.String(), cid, "--from", key.name)
	args = append(args, moreflags...)
	return execCmdWithInput(args, KeyPwd)
}

func (cli Emcli) MarketCancelOrder(key Key, cid string) (string, bool, error) {
	args := cli.addTransactionFlags("tx", "market", "cancel", cid, "--from", key.name)
	return execCmdWithInput(args, KeyPwd)
}

func (cli Emcli) UnjailValidator(key string) (string, bool, error) {
	args := cli.addTransactionFlags("tx", "slashing", "unjail", "--from", key)
	return execCmdWithInput(args, KeyPwd)
}

func (cli Emcli) BEP3Create(creator Key, recipient, otherChainRecipient, otherChainSender, coins string, TTL int) (string, string, string, error) {
	args := cli.addTransactionFlags("tx", "bep3", "create", recipient, otherChainRecipient, otherChainSender, "now", coins, fmt.Sprint(TTL), "--from", creator.name)
	output, err := execCmdCollectOutput(args, KeyPwd)
	if err != nil {
		return "", "", "", err
	}

	re := regexp.MustCompile("(?i)(Random number: (?P<randomnumber>\\w+)|Timestamp: (?P<timestamp>\\d+)|Random number hash: (?P<randomnumberhash>\\w+))")
	groups := extractNamedGroups(output, re)

	var (
		randNumber     = groups["randomnumber"]
		randNumberHash = groups["randomnumberhash"]
		timestamp      = groups["timestamp"]
	)

	return randNumber, randNumberHash, timestamp, nil
}

func (cli Emcli) BEP3Claim(claimant Key, swapId, secret string) (string, error) {
	args := cli.addTransactionFlags("tx", "bep3", "claim", swapId, secret, "--from", claimant.name)

	return execCmdCollectOutput(args, KeyPwd)
}

func extractTxHash(bz []byte) (txhash string, success bool, err error) {
	json := gjson.ParseBytes(bz)

	txhashjson := json.Get("txhash")
	logs := json.Get("logs")
	code := json.Get("code")

	// todo (reviewer) : emd command returns `exit 0` although the TX has failed with `signature verification failed`
	// any non zero `code` in response json is a failure code
	if !txhashjson.Exists() || !logs.Exists() || code.Int() != 0 {
		return "", false, fmt.Errorf("tx appears to have failed %v", string(bz))
	}

	return txhashjson.Str, true, nil
}

func execCmdCollectOutput(arguments []string, input string) (string, error) {
	cmd := exec.Command(EMCLI, arguments...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}

	_, err = io.WriteString(stdin, input+"\n")
	if err != nil {
		return "", err
	}

	// fmt.Println(" *** Running command: ", EMCLI, strings.Join(arguments, " "))
	// bz, err := cmd.CombinedOutput()
	var b bytes.Buffer
	cmd.Stderr = &b

	bz, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(bz), nil
}

func execCmdWithInput(arguments []string, input string) (string, bool, error) {
	cmd := exec.Command(EMCLI, arguments...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", false, err
	}

	_, err = io.WriteString(stdin, input+"\n")
	if err != nil {
		return "", false, err
	}

	// fmt.Println(" *** Running command: ", EMCLI, strings.Join(arguments, " "))
	bz, err := cmd.CombinedOutput()
	// fmt.Println(" *** CombinedOutput", string(bz))
	if err != nil {
		return "", false, err
	}

	return extractTxHash(bz)
}

func execCmdAndCollectResponse(arguments []string) ([]byte, error) {
	// fmt.Println(" *** Running command: ", EMCLI, strings.Join(arguments, " "))
	bz, err := exec.Command(EMCLI, arguments...).CombinedOutput()
	// fmt.Println(" *** Output: ", string(bz))
	return bz, err
}

func (cli Emcli) addQueryFlags(arguments ...string) []string {
	arguments = append(arguments, "--output", "json")
	return cli.addNetworkFlags(arguments)
}

func (cli Emcli) addTransactionFlags(arguments ...string) []string {
	arguments = append(arguments,
		"--home", cli.keystore.path,
		"--keyring-backend", "test",
		"--broadcast-mode", "block",
		"--yes",
	)

	return cli.addNetworkFlags(arguments)
}

func (cli Emcli) addNetworkFlags(arguments []string) []string {
	return append(arguments,
		"--node", cli.node,
		"--chain-id", cli.chainid,
	)
}

func extractNamedGroups(input string, re *regexp.Regexp) map[string]string {
	groupNames := re.SubexpNames()
	result := make(map[string]string)

	for _, match := range re.FindAllStringSubmatch(input, -1) {
		for groupIdx, group := range match {
			if groupNames[groupIdx] == "" || len(group) == 0 {
				continue
			}

			result[groupNames[groupIdx]] = group
		}
	}

	return result
}
