package emoney

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	authsign "github.com/cosmos/cosmos-sdk/x/auth/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/e-money/em-ledger/x/authority"
	"github.com/e-money/em-ledger/x/market/types"
	"github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	abci "github.com/tendermint/tendermint/abci/types"
)

func setupMarketApp(t *testing.T, options ...func(*baseapp.BaseApp)) (sdk.Context, *EMoneyApp, EncodingConfig) {
	t.Helper()

	encCfg := MakeEncodingConfig()

	db := dbm.NewMemDB()

	app := NewApp(
		log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true,
		map[int64]bool{}, t.TempDir(), 0, encCfg, EmptyAppOptions{}, options...,
	)
	require.NotNil(t, app)

	ctx := app.BaseApp.NewUncachedContext(true, tmproto.Header{ChainID: "test-market-chain"})
	ctx = ctx.WithBlockTime(time.Now())

	app.marketKeeper.InitParamsStore(ctx)

	return ctx, app, encCfg
}

func coins(s string) sdk.Coins {
	coins, err := sdk.ParseCoinsNormalized(s)
	if err != nil {
		panic(err)
	}
	return coins
}

func TestAppLimitOrder_0_Full_Err_Gas(t *testing.T) {
	ctx, app, enc := setupMarketApp(t,  []func(*baseapp.BaseApp){}...)
	require.NotNil(t, app)

	genesisState := ModuleBasics.DefaultGenesis(enc.Marshaler)
	authorityState := authority.NewGenesisState(
		rand.Bytes(sdk.AddrLen), nil,
		sdk.NewDecCoins(),
	)
	genesisState[authority.ModuleName] = enc.Marshaler.MustMarshalJSON(&authorityState)

	keystore1, acct1 := getNewAcctInfo(t)
	keystore2, acct2 := getNewAcctInfo(t)

	bal1 := coins("100000chf,1000000ngm")
	bal2 := coins("100000eur,1000000ngm")
	supply := coins("100000chf,100000eur,2000000ngm")

	bankState := banktypes.NewGenesisState(
		banktypes.DefaultParams(),
		[]banktypes.Balance{
			{acct1.GetAddress().String(), bal1},
			{acct2.GetAddress().String(), bal2},
		},
		supply,
		[]banktypes.Metadata{},
	)
	genesisState[banktypes.ModuleName] = enc.Marshaler.MustMarshalJSON(bankState)

	stateBytes, err := json.MarshalIndent(genesisState, "", "  ")
	require.NoError(t, err)

	app.bankKeeper.SetSupply(ctx, banktypes.NewSupply(supply))

	// Initialize the chain
	app.InitChain(
		abci.RequestInitChain{
			Validators:    []abci.ValidatorUpdate{},
			AppStateBytes: stateBytes,
		},
	)

	total := app.bankKeeper.GetSupply(ctx)
	require.NotNil(t, total)

	app.accountKeeper.SetAccount(
		ctx, app.accountKeeper.NewAccountWithAddress(ctx, acct1.GetAddress()),
	)
	app.accountKeeper.SetAccount(
		ctx, app.accountKeeper.NewAccountWithAddress(ctx, acct2.GetAddress()),
	)

	//
	// Liquid 0 Gas Cost
	//
	msg := &types.MsgAddLimitOrder{
		TimeInForce:   types.TimeInForce_GoodTillCancel,
		Owner:         acct1.GetAddress().String(),
		Source:        sdk.NewCoin("chf", sdk.NewInt(50000)),
		Destination:   sdk.NewCoin("eur", sdk.NewInt(50000)),
		ClientOrderId: "testAddLimitOrder-chf-eur1",
	}

	tx := getSignedTrx(ctx, t, app.accountKeeper, enc, msg, keystore1, acct1, 0, 0)

	gasInfo, _, err := app.Deliver(enc.TxConfig.TxEncoder(), tx)
	require.NoError(t, err)
	require.Equal(t, gasInfo.GasUsed, sdk.Gas(0))

	//
	// Destination denomination xxx does not exist and errs, full gas
	//
	msg2 := &types.MsgAddLimitOrder{
		TimeInForce:   types.TimeInForce_GoodTillCancel,
		Owner:         acct1.GetAddress().String(),
		Source:        sdk.NewCoin("chf", sdk.NewInt(50000)),
		Destination:   sdk.NewCoin("xxx", sdk.NewInt(50000)),
		ClientOrderId: "testAddLimitOrder-chf-eur2",
	}

	tx = getSignedTrx(ctx, t, app.accountKeeper, enc, msg2, keystore1, acct1, 0, 1)

	gasInfo, _, err = app.Deliver(enc.TxConfig.TxEncoder(), tx)
	require.Error(t, err)
	require.Equal(t, gasInfo.GasUsed, sdk.Gas(25000))

	msg3 := &types.MsgAddLimitOrder{
		TimeInForce:   types.TimeInForce_GoodTillCancel,
		Owner:         acct2.GetAddress().String(),
		Source:        sdk.NewCoin("eur", sdk.NewInt(50000)),
		Destination:   sdk.NewCoin("chf", sdk.NewInt(50000)),
		ClientOrderId: "testAddLimitOrder-eur-chf",
	}

	tx = getSignedTrx(ctx, t, app.accountKeeper, enc, msg3, keystore2, acct2, 0, 0)

	gasInfo, _, err = app.Deliver(enc.TxConfig.TxEncoder(), tx)
	require.NoError(t, err)
	require.Equal(t, gasInfo.GasUsed, sdk.Gas(25000))
}

func getSignedTrx(
	ctx sdk.Context,
	t *testing.T, ak types.AccountKeeper, enc EncodingConfig,
	msg *types.MsgAddLimitOrder,
	keystore keyring.Keyring, acct keyring.Info,
	accountNumber, sequence uint64,
) authsign.Tx {
	acci := ak.GetAccount(ctx, acct.GetAddress())
	require.Equal(t, acci.GetAddress().String(), acct.GetAddress().String())

	txBuilder := enc.TxConfig.NewTxBuilder()
	txBuilder.SetMsgs(msg)
	txBuilder.SetFeeAmount(coins("25000ngm"))
	txBuilder.SetGasLimit(213456)
	txBuilder.SetMemo("TestMarketOrder")

	txFactory := clienttx.Factory{}.
		WithChainID("").
		WithTxConfig(enc.TxConfig).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT).
		WithAccountNumber(accountNumber).
		WithSequence(sequence).
		WithKeybase(keystore)

	tx := txBuilder.GetTx()
	signers := tx.GetSigners()
	require.Equal(t, signers[0].String(), acct.GetAddress().String())

	err := authclient.SignTx(
		txFactory, client.Context{}, acct.GetName(),
		txBuilder, true, true,
	)
	require.NoError(t, err)

	_, err = enc.TxConfig.TxEncoder()(txBuilder.GetTx())
	require.NoError(t, err)

	return tx
}

func getNewAcctInfo(t *testing.T) (keyring.Keyring, keyring.Info) {
	keystore, err := keyring.New(t.Name()+"1", keyring.BackendMemory, "", nil)
	require.NoError(t, err)

	uid := "theKeyName"

	info, _, err := keystore.NewMnemonic(
		uid, keyring.English, sdk.FullFundraiserPath, hd.Secp256k1,
	)
	require.NoError(t, err)

	return keystore, info
}

func TestSimAppExportAndBlockedAddrs(t *testing.T) {
	encCfg := MakeEncodingConfig()
	db := dbm.NewMemDB()
	app := NewApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, map[int64]bool{}, t.TempDir(), 0, encCfg, EmptyAppOptions{})

	for acc := range maccPerms {
		require.True(
			t,
			app.bankKeeper.BlockedAddr(app.accountKeeper.GetModuleAddress(acc)),
			"ensure that blocked addresses are properly set in bank keeper",
		)
	}

	genesisState := ModuleBasics.DefaultGenesis(encCfg.Marshaler)
	authorityState := authority.NewGenesisState(rand.Bytes(sdk.AddrLen), nil, sdk.NewDecCoins())
	genesisState[authority.ModuleName] = encCfg.Marshaler.MustMarshalJSON(&authorityState)

	stateBytes, err := json.MarshalIndent(genesisState, "", "  ")
	require.NoError(t, err)

	// Initialize the chain
	app.InitChain(
		abci.RequestInitChain{
			Validators:    []abci.ValidatorUpdate{},
			AppStateBytes: stateBytes,
		},
	)
	app.Commit()

	// Making a new app object with the db, so that initchain hasn't been called
	app2 := NewApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, map[int64]bool{}, t.TempDir(), 0, encCfg, EmptyAppOptions{})
	_, err = app2.ExportAppStateAndValidators(false, []string{})
	require.NoError(t, err, "ExportAppStateAndValidators should not have an error")
}

// EmptyAppOptions is a stub implementing AppOptions
type EmptyAppOptions struct{}

// Get implements AppOptions
func (ao EmptyAppOptions) Get(o string) interface{} {
	return nil
}
