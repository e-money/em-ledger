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

func TestAppLiquidOrder0Gas(t *testing.T) {
	ctx, app, enc := setupMarketApp(t,  []func(*baseapp.BaseApp){}...)
	require.NotNil(t, app)

	genesisState := ModuleBasics.DefaultGenesis(enc.Marshaler)
	authorityState := authority.NewGenesisState(
		rand.Bytes(sdk.AddrLen), nil,
		sdk.NewDecCoins(),
	)
	genesisState[authority.ModuleName] = enc.Marshaler.MustMarshalJSON(&authorityState)

	keystore, err := keyring.New(t.Name(), keyring.BackendMemory, "", nil)
	require.NoError(t, err)

	uid := "theKeyName"

	info, _, err := keystore.NewMnemonic(
		uid, keyring.English, sdk.FullFundraiserPath, hd.Secp256k1,
	)
	require.NoError(t, err)
	t.Log(info.GetAddress().String())

	supply := coins("100000chf,100000eur,1000000ngm")

	bankState := banktypes.NewGenesisState(
		banktypes.DefaultParams(),
		[]banktypes.Balance{{info.GetAddress().String(), supply}},
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
		ctx, app.accountKeeper.NewAccountWithAddress(ctx, info.GetAddress()),
	)
	acci := app.accountKeeper.GetAccount(ctx, info.GetAddress())
	t.Log(acci.GetAddress().String())
	require.NotNil(t, acci)
	require.Equal(t, acci.GetAddress().String(), info.GetAddress().String())

	err = app.bankKeeper.SetBalances(
		ctx, info.GetAddress(),	coins("250000ngm,50000chf"))
	require.NoError(t, err)

	msg := &types.MsgAddLimitOrder{
		TimeInForce:   types.TimeInForce_GoodTillCancel,
		Owner:         info.GetAddress().String(),
		Source:        sdk.NewCoin("chf", sdk.NewInt(50000)),
		Destination:   sdk.NewCoin("eur", sdk.NewInt(60000)),
		ClientOrderId: "testAddLimitOrder-chf-eur",
	}

	txBuilder := enc.TxConfig.NewTxBuilder()
	txBuilder.SetMsgs(msg)
	txBuilder.SetFeeAmount(coins("25000ngm"))
	txBuilder.SetGasLimit(213456)
	txBuilder.SetMemo("TestAppLiquidOrder0Gas")

	txFactory := clienttx.Factory{}.
		WithChainID("").
		WithTxConfig(enc.TxConfig).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT).
		WithAccountNumber(acci.GetAccountNumber()).
		WithSequence(acci.GetSequence()).
		WithKeybase(keystore)

	tx := txBuilder.GetTx()
	signers := tx.GetSigners()
	require.Equal(t, signers[0].String(), info.GetAddress().String())

	err = authclient.SignTx(txFactory, client.Context{}, info.GetName(),
		txBuilder, true, true)
	require.NoError(t, err)

	_, err = enc.TxConfig.TxEncoder()(txBuilder.GetTx())
	require.NoError(t, err)

	gasInfo, _, err := app.Deliver(enc.TxConfig.TxEncoder(), tx)
	require.NoError(t, err)
	require.Equal(t, gasInfo.GasUsed, sdk.Gas(0))
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
