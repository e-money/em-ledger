// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

// +build bdd

package networktest

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
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

func CreateMultiMsgTx(key Key, chainid, feestring string, accnum, sequence uint64, msgs ...sdk.Msg) types.StdTx {
	memo := "+memo"
	fee := auth.NewStdFee(100000, mustParseCoins(feestring))

	sig, err := key.Sign(
		auth.StdSignBytes(
			chainid,
			accnum,
			sequence,
			fee,
			msgs,
			memo),
	)

	if err != nil {
		panic(err)
	}

	signature := auth.StdSignature{
		PubKey:    key.pubkey,
		Signature: sig,
	}

	tx := auth.NewStdTx(msgs, fee, []auth.StdSignature{signature}, memo)
	return tx
}

func mustParseCoins(coins string) sdk.Coins {
	cs, err := sdk.ParseCoins(coins)
	if err != nil {
		panic(err)
	}

	return cs
}

func mustParseCoin(coin string) sdk.Coin {
	c, err := sdk.ParseCoin(coin)
	if err != nil {
		panic(err)
	}

	return c
}
