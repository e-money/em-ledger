// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

// +build bdd

package emoney_test

import (
	"bytes"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	emoney "github.com/e-money/em-ledger"
	nt "github.com/e-money/em-ledger/networktest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/pflag"
	rpcclient "github.com/tendermint/tendermint/rpc/client/http"
	"github.com/tidwall/gjson"
	"os"
	"sync"
	"sync/atomic"
)

var _ = Describe("Staking", func() {

	var (
		Key1 = testnet.Keystore.Key1
		Key2 = testnet.Keystore.Key2
		Key3 = testnet.Keystore.Key3
	)

	Describe("Blocks can hold many transactions", func() {
		Context("", func() {
			It("creates a new testnet", createNewTestnet)

			It("Creates a lot of send transactions", func() {
				const trxCount = 400

				var (
					failedTxs int32 = 0
					coin, _         = sdk.ParseCoinsNormalized("15000eeur")
					chainID         = testnet.ChainID()
					txhash          = make(chan string, 1024)
				)

				senders := []nt.Key{Key1, Key2, Key3, Key3}
				receivers := []nt.Key{Key2, Key1, Key1, Key2}

				for i := 0; i < trxCount; i++ {

					go func(from, to nt.Key) {
						hash, err := sendTx(from, to, coin, chainID)
						if err != nil {
							atomic.AddInt32(&failedTxs, 1)
							fmt.Println(err)
							return
						}

						txhash <- hash
					}(senders[i%4], receivers[i%4])
				}

				_, _ = nt.IncChain(1)

				emcli := testnet.NewEmcli()
				success, failure, errs := 0, 0, failedTxs
				for ; success+failure+int(errs) < trxCount;{
					h := <-txhash
					bz, err := emcli.QueryTransaction(h)
					if err != nil {
						errs++
						continue
					}

					s := gjson.ParseBytes(bz).Get("txhash")
					if s.Exists() {
						success++
					} else {
						failure++
					}
				}

				fmt.Printf(" *** Transactions summary:\n Successful: %v\n Failed: %v\n Errors: %v\n Total: %v\n", success, failure, errs, success+failure+int(errs))
				Expect(success).To(Equal(trxCount))

				// Equivalent event listening handler
				// listener, err := nt.NewEventListener()
				// Expect(err).ToNot(HaveOccurred())
				//
				// Eventually(func() int {
				//	 success, err = listener.SubTx(
				//	 	mu, txHashes, trxCount-errs, expiration,
				//	 )
				//	 return int(success)
				// }, expiration, time.Second).Should(Equal(trxCount))
			})
		})
	})
})

type accountNoSequence struct {
	AccountNo, Sequence uint64
}

var (
	sendMutex sync.Mutex
	sequences = make(map[string]accountNoSequence)
)

func sendTx(fromKey, toKey nt.Key, amount sdk.Coins, chainID string) (string, error) {
	sendMutex.Lock()
	defer sendMutex.Unlock()

	from, err := sdk.AccAddressFromBech32(fromKey.GetAddress())
	if err != nil {
		return "", err
	}

	encodingConfig := emoney.MakeEncodingConfig()

	httpClient, err := rpcclient.New("tcp://localhost:26657", "/websocket")
	if err != nil {
		return "", err
	}

	clientCtx := client.Context{}.
		WithJSONMarshaler(encodingConfig.Marshaler).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(authtypes.AccountRetriever{}).
		WithBroadcastMode(flags.BroadcastAsync).
		WithHomeDir(emoney.DefaultNodeHome).
		WithChainID(chainID).
		WithFromName(fromKey.GetName()).
		WithFromAddress(from).
		WithKeyring(testnet.Keystore.Keyring()).
		WithClient(httpClient).
		WithSkipConfirmation(true)

	var (
		accInfo accountNoSequence
		present bool
	)
	if accInfo, present = sequences[fromKey.GetAddress()]; !present {
		accountNumber, sequence, err := authtypes.AccountRetriever{}.GetAccountNumberSequence(clientCtx, from)
		if err != nil {
			return "", err
		}

		accInfo = accountNoSequence{
			AccountNo: accountNumber,
			Sequence:  sequence,
		}
	}

	sendMsg := &banktypes.MsgSend{
		FromAddress: fromKey.GetAddress(),
		ToAddress:   toKey.GetAddress(),
		Amount:      amount,
	}
	if err := sendMsg.ValidateBasic(); err != nil {
		return "", err
	}
	flagSet := pflag.NewFlagSet("testing", pflag.PanicOnError)
	txf := tx.NewFactoryCLI(clientCtx, flagSet).
		WithMemo("+memo").
		WithSequence(accInfo.Sequence).
		WithAccountNumber(accInfo.AccountNo)

	accInfo.Sequence++
	sequences[fromKey.GetAddress()] = accInfo

	var buf bytes.Buffer
	err = tx.BroadcastTx(clientCtx.WithOutput(&buf), txf, sendMsg)
	if err != nil {
		return "", err
	}
	var resp sdk.TxResponse
	return resp.TxHash, encodingConfig.Marshaler.UnmarshalJSON(buf.Bytes(), &resp)
}
