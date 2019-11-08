// +build bdd

package emoney

import (
	"fmt"
	"sync"
	"time"

	nt "emoney/networktest"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	atypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	rpcclient "github.com/tendermint/tendermint/rpc/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
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
						coins, _ := sdk.ParseCoins("15000x2eur")
						txResponse, err := sendTx(Key1, Key2, coins, testnet.ChainID())
						if err != nil {
							fmt.Println(err)
						}

						txhash <- txResponse.TxHash
					}()

					go func() {
						coins, _ := sdk.ParseCoins("9000x2eur")
						txResponse, err := sendTx(Key2, Key1, coins, testnet.ChainID())
						if err != nil {
							fmt.Println(err)
						}

						txhash <- txResponse.TxHash
					}()

					go func() {
						coins, _ := sdk.ParseCoins("4000x2eur")
						txResponse, err := sendTx(Key3, Key1, coins, testnet.ChainID())
						if err != nil {
							fmt.Println(err)
						}

						txhash <- txResponse.TxHash
					}()

					go func() {
						coins, _ := sdk.ParseCoins("7700x2eur")
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

					s := gjson.ParseBytes(bz).Get("logs.0.success")
					if s.Bool() {
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

	cdc := codec.New()
	atypes.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	bank.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	cliCtx := context.NewCLIContext().
		WithCodec(cdc).
		//WithBroadcastMode("block").
		WithBroadcastMode("async").
		WithClient(rpcclient.NewHTTP("tcp://localhost:26657", "/websocket"))

	to, err := sdk.AccAddressFromBech32(toKey.GetAddress())
	if err != nil {
		return sdk.TxResponse{}, err
	}

	from, err := sdk.AccAddressFromBech32(fromKey.GetAddress())
	if err != nil {
		return sdk.TxResponse{}, err
	}

	var (
		accInfo accountNoSequence
		present bool
	)
	if accInfo, present = sequences[fromKey.GetAddress()]; !present {
		accountNumber, sequence, err := atypes.NewAccountRetriever(cliCtx).GetAccountNumberSequence(from)
		if err != nil {
			return sdk.TxResponse{}, err
		}

		accInfo = accountNoSequence{
			AccountNo: accountNumber,
			Sequence:  sequence,
		}
	}

	sendMsg := bank.MsgSend{
		FromAddress: from,
		ToAddress:   to,
		Amount:      amount,
	}

	txBldr := auth.NewTxBuilderFromCLI().
		WithTxEncoder(utils.GetTxEncoder(cdc)).
		WithChainID(chainID).
		WithAccountNumber(accInfo.AccountNo).
		WithSequence(accInfo.Sequence)

	accInfo.Sequence++

	sequences[fromKey.GetAddress()] = accInfo

	signMsg, err := txBldr.BuildSignMsg([]sdk.Msg{sendMsg})
	if err != nil {
		return sdk.TxResponse{}, err
	}

	sigBytes, err := fromKey.Sign(signMsg.Bytes())
	if err != nil {
		return sdk.TxResponse{}, err
	}

	stdSignature := atypes.StdSignature{
		PubKey:    fromKey.GetPublicKey(),
		Signature: sigBytes,
	}

	tx := auth.NewStdTx(signMsg.Msgs, signMsg.Fee, []atypes.StdSignature{stdSignature}, signMsg.Memo)

	txBytes, err := txBldr.TxEncoder()(tx)
	if err != nil {
		return sdk.TxResponse{}, err
	}

	return cliCtx.BroadcastTx(txBytes)
}
