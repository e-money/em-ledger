// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

// +build bdd

package networktest

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	emoney "github.com/e-money/em-ledger"
	"strings"
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
		if strings.Contains(s, substring) {
			scanOnce.Do(mutex.Unlock)
		}
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
	ec := emoney.MakeEncodingConfig()

	clientCtx := client.Context{
		FromAddress:       key.address,
		ChainID:           chainid,
		JSONMarshaler:     ec.Marshaler,
		InterfaceRegistry: ec.InterfaceRegistry,
		Keyring:           key.keybase,
		From:              key.name,
		FromName:          key.name,
		TxConfig:          ec.TxConfig,
		LegacyAmino:       ec.Amino,
	}
	txf := tx.NewFactoryCLI(clientCtx, nil)
	txf.WithMemo("+memo").WithFees(feestring).WithSequence(sequence).WithAccountNumber(accnum)

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
