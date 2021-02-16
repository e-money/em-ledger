// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

// +build bdd

package emoney_test

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	atypes "github.com/cosmos/cosmos-sdk/x/auth/types"
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
	"time"
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
				emcli := testnet.NewEmcli()

				time.Sleep(2 * time.Second)

				txhash := make(chan string, 1024)

				for i := 0; i < 100; i++ {
					go func() {
						coins, _ := sdk.ParseCoinsNormalized("15000eeur")
						txResponse, err := sendTx(Key1, Key2, coins, testnet.ChainID())
						if err != nil {
							fmt.Println(err)
						}

						txhash <- txResponse.TxHash
					}()

					go func() {
						coins, _ := sdk.ParseCoinsNormalized("9000eeur")
						txResponse, err := sendTx(Key2, Key1, coins, testnet.ChainID())
						if err != nil {
							fmt.Println(err)
						}

						txhash <- txResponse.TxHash
					}()

					go func() {
						coins, _ := sdk.ParseCoinsNormalized("4000eeur")
						txResponse, err := sendTx(Key3, Key1, coins, testnet.ChainID())
						if err != nil {
							fmt.Println(err)
						}

						txhash <- txResponse.TxHash
					}()

					go func() {
						coins, _ := sdk.ParseCoinsNormalized("7700eeur")
						txResponse, err := sendTx(Key3, Key2, coins, testnet.ChainID())
						if err != nil {
							fmt.Println(err)
						}

						txhash <- txResponse.TxHash
					}()
				}

				txHashes := make([]string, 0)

				go func() {
					for h := range txhash {
						txHashes = append(txHashes, h)
					}
				}()

				time.Sleep(30 * time.Second)

				success, failure, errs := 0, 0, 0
				for _, h := range txHashes {
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

				fmt.Printf(" *** Transactions summary:\n Successful: %v\n Failed: %v\n Errors: %v\n Total: %v\n", success, failure, errs, success+failure+errs)
				Expect(success).To(Equal(400))
			})
		})
	})
})

var (
	sendMutex sync.Mutex
)

type accountNoSequence struct {
	AccountNo, Sequence uint64
}

var (
	sequences map[string]accountNoSequence = make(map[string]accountNoSequence)
)

func sendTx(fromKey, toKey nt.Key, amount sdk.Coins, chainID string) (sdk.TxResponse, error) {
	sendMutex.Lock()
	defer sendMutex.Unlock()

	encodingConfig := emoney.MakeEncodingConfig()

	httpClient, err := rpcclient.New("tcp://localhost:26657", "/websocket")
	if err != nil {
		return sdk.TxResponse{}, err
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
		WithFrom(fromKey.GetAddress()).
		WithKeyring(testnet.Keystore.Keyring()).
		WithClient(httpClient)

	var (
		accInfo accountNoSequence
		present bool
	)
	if accInfo, present = sequences[fromKey.GetAddress()]; !present {
		accountNumber, sequence, err := atypes.NewAccountRetriever(clientCtx).GetAccountNumberSequence(from)
		if err != nil {
			return sdk.TxResponse{}, err
		}

		accInfo = accountNoSequence{
			AccountNo: accountNumber,
			Sequence:  sequence,
		}
	}

	sendMsg := banktypes.MsgSend{
		FromAddress: fromKey.GetAddress(),
		ToAddress:   toKey.GetAddress(),
		Amount:      amount,
	}
	flagSet := pflag.NewFlagSet("testing", pflag.PanicOnError)
	txf := tx.NewFactoryCLI(clientCtx, flagSet)
	txf.WithMemo("+memo").WithSequence(accInfo.Sequence).WithAccountNumber(accInfo.AccountNo)

	accInfo.Sequence++

	sequences[fromKey.GetAddress()] = accInfo

	txb, err := tx.BuildUnsignedTx(txf, sendMsg)
	if err != nil {
		return sdk.TxResponse{}, err
	}
	//err = tx.Sign(txf, fromKey.GetAddress(), txb, false)
	if err != nil {
		return sdk.TxResponse{}, err
	}

	return tx.BroadcastTx(clientCtx, txf, sendMsg)
}
