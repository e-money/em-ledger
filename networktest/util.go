// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

// +build bdd

package networktest

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	emoney "github.com/e-money/em-ledger"
	"github.com/spf13/pflag"
	"os"
	"sync"
	"time"
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

func CreateMultiMsgTx(key Key, chainid, feestring string, accnum, sequence uint64, msgs ...sdk.Msg) signing.Tx {
	for i, m := range msgs {
		if err := m.ValidateBasic(); err != nil {
			panic(fmt.Sprintf("invalid msg at pos %d: %#v", i, m))
		}
	}
	encodingConfig := emoney.MakeEncodingConfig()

	clientCtx := client.Context{}.
		WithJSONMarshaler(encodingConfig.Marshaler).
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
