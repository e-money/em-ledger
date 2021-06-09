// +build bdd

package networktest

import (
	"bytes"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	emoney "github.com/e-money/em-ledger"
	"github.com/spf13/pflag"
	rpcclient "github.com/tendermint/tendermint/rpc/client/http"
	"os"
	"sync"
)

type accountNoSequence struct {
	AccountNo, Sequence uint64
}

var (
	sendMutex sync.Mutex
	sequences = make(map[string]accountNoSequence)
)

func (t Testnet) SendTx(fromKey, toKey Key, amount sdk.Coins, chainID string) (string, error) {
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
		WithKeyring(t.Keystore.Keyring()).
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

