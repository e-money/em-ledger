//go:build bdd
// +build bdd

package networktest

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/e-money/em-ledger/x/issuer/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	emoney "github.com/e-money/em-ledger"
	"github.com/spf13/pflag"
)

// Create a scanner function with built-in timeout. The returned wait function blocks until the
// substring has been encountered or the provided timeout has been reached.
// Results from more than one invocation of the returned wait-function are undefined.
func createOutputScanner(substring string, timeout time.Duration) (wait func() bool, scanner func(string)) {
	mutex := &sync.Mutex{}
	mutex.Lock()
	scanOnce := sync.Once{}

	scanner = func(s string) {
		scanOnce.Do(mutex.Unlock)
	}

	// Bridge mutex to a regular channel
	fn := func() <-chan interface{} {
		c := make(chan interface{}, 0)

		go func() {
			mutex.Lock()
			c <- true
		}()

		return c
	}

	wait = func() bool {
		select {
		case <-time.Tick(timeout):
			return false
		case <-fn():
			return true
		}
	}

	return
}

func AuthCreatesIssuer(emcli Emcli, Authority, Issuer Key) bool {
	issuers, denoms := CreateIssuer(emcli, Authority, Issuer, "eeur", `'ejpy,eJPY,Japanese yen stablecoin'`)

	return len(issuers) == 1 && strings.Contains(
		denoms, "eeur",
	) && strings.Contains(denoms, "ejpy")
}

func CreateIssuer(emcli Emcli, Authority Key, Issuer Key, denomArgs ...string) ([]types.Issuer, string) {
	_, success, err := emcli.AuthorityCreateIssuer(
		Authority, Issuer, denomArgs...,
	)
	if err != nil || !success {
		return nil, ""
	}

	bz, err := emcli.QueryIssuers()
	if err != nil {
		return nil, ""
	}

	var resp types.QueryIssuersResponse
	if err = json.Unmarshal(bz, &resp); err != nil {
		return nil, ""
	}

	return resp.Issuers, strings.Join(resp.Issuers[len(resp.Issuers)-1].Denoms, ",")
}

func CreateMultiMsgTx(key Key, chainid, feestring string, accnum, sequence uint64, msgs ...sdk.Msg) signing.Tx {
	for i, m := range msgs {
		if err := m.ValidateBasic(); err != nil {
			panic(fmt.Sprintf("invalid msg at pos %d: %#v", i, m))
		}
	}
	encodingConfig := emoney.MakeEncodingConfig()

	clientCtx := client.Context{}.
		WithJSONCodec(encodingConfig.Marshaler).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(authtypes.AccountRetriever{}).
		WithBroadcastMode(flags.BroadcastBlock).
		WithHomeDir(emoney.DefaultNodeHome).
		WithChainID(chainid).
		WithFrom(key.address.String()).
		WithKeyring(key.keybase)

	flagSet := pflag.NewFlagSet("testing", pflag.PanicOnError)
	txf := tx.NewFactoryCLI(clientCtx, flagSet).
		WithMemo("+memo").
		WithFees(feestring).
		WithSequence(sequence).
		WithAccountNumber(accnum)

	txb, err := tx.BuildUnsignedTx(txf, msgs...)
	if err != nil {
		panic("failed to build tx: " + err.Error())
	}
	err = tx.Sign(txf, key.name, txb, false)
	if err != nil {
		panic("failed to sign tx: " + err.Error())
	}

	return txb.GetTx()
}
