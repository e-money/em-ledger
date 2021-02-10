// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	apptypes "github.com/e-money/em-ledger/types"
	"github.com/e-money/em-ledger/x/authority/types"
	"github.com/e-money/em-ledger/x/issuer"
	"github.com/e-money/em-ledger/x/liquidityprovider"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

func init() {
	// Be able to parse emoney bech32 encoded addresses.
	apptypes.ConfigureSDK()
}

func TestAuthorityBasicPersistence(t *testing.T) {
	ctx, keeper, _, _ := createTestComponents(t)

	require.Panics(t, func() {
		// Keeper must panic if no authority has been specified
		keeper.GetAuthority(ctx)
	})

	acc, _ := sdk.AccAddressFromBech32("emoney1kt0vh0ttget0xx77g6d3ttnvq2lnxx6vp3uyl0")
	keeper.SetAuthority(ctx, acc)

	authority := keeper.GetAuthority(ctx)
	require.Equal(t, acc, authority)
}

func TestMustBeAuthority(t *testing.T) {
	ctx, keeper, _, _ := createTestComponents(t)

	var (
		accAuthority = mustParseAddress("emoney1kt0vh0ttget0xx77g6d3ttnvq2lnxx6vp3uyl0")
		acc2         = mustParseAddress("emoney17up20gamd0vh6g9ne0uh67hx8xhyfrv2lyazgu")
	)

	require.Panics(t, func() {
		// Must panic due to authority not being set yet.
		keeper.MustBeAuthority(ctx, accAuthority)
	})

	keeper.SetAuthority(ctx, accAuthority)
	keeper.MustBeAuthority(ctx, accAuthority)

	require.Panics(t, func() {
		keeper.MustBeAuthority(ctx, acc2)
	})

	// Authority can only be specified once, preferably during genesis
	require.Panics(t, func() {
		keeper.SetAuthority(ctx, acc2)
	})
}

func TestCreateAndRevokeIssuer(t *testing.T) {
	ctx, keeper, ik, _ := createTestComponents(t)

	var (
		accAuthority = mustParseAddress("emoney1kt0vh0ttget0xx77g6d3ttnvq2lnxx6vp3uyl0")
		issuer1      = mustParseAddress("emoney17up20gamd0vh6g9ne0uh67hx8xhyfrv2lyazgu")
		issuer2      = mustParseAddress("emoney1dgkjvr2kkrp0xc5qn66g23us779q2dmgle5aum")
	)

	keeper.SetAuthority(ctx, accAuthority)

	_, err := keeper.CreateIssuer(ctx, accAuthority, issuer1, []string{"eeur", "ejpy"})
	require.NoError(t, err)

	_, err = keeper.CreateIssuer(ctx, accAuthority, issuer2, []string{"echf", "egbp", "eeur"})
	require.Error(t, err) // Must fail due to duplicate token denomination

	_, err = keeper.CreateIssuer(ctx, accAuthority, issuer2, []string{"echf", "egbp"})
	require.NoError(t, err)
	require.Len(t, ik.GetIssuers(ctx), 2)

	_, err = keeper.DestroyIssuer(ctx, accAuthority, issuer2)
	require.NoError(t, err)
	require.Len(t, ik.GetIssuers(ctx), 1)

	require.Panics(t, func() {
		// Make sure only authority key can destroy an issuer
		keeper.DestroyIssuer(ctx, issuer1, issuer2)
	})

	_, err = keeper.DestroyIssuer(ctx, accAuthority, issuer2)
	require.Error(t, err)
	require.Len(t, ik.GetIssuers(ctx), 1)

	_, err = keeper.DestroyIssuer(ctx, accAuthority, issuer1)
	require.NoError(t, err)
	require.Empty(t, ik.GetIssuers(ctx))
}

func TestAddMultipleDenomsSameIssuer(t *testing.T) {
	ctx, keeper, ik, _ := createTestComponents(t)

	var (
		accAuthority = mustParseAddress("emoney1kt0vh0ttget0xx77g6d3ttnvq2lnxx6vp3uyl0")
		accIssuer    = mustParseAddress("emoney17up20gamd0vh6g9ne0uh67hx8xhyfrv2lyazgu")
	)

	keeper.SetAuthority(ctx, accAuthority)

	_, err := keeper.CreateIssuer(ctx, accAuthority, accIssuer, []string{"eeur", "ejpy"})
	require.NoError(t, err)

	_, err = keeper.CreateIssuer(ctx, accAuthority, accIssuer, []string{"ekrw"})
	require.NoError(t, err)
	issuers := ik.GetIssuers(ctx)

	// Ensure that the denomination has been added to the existing issuer, not to a new entry with the same key
	require.Len(t, issuers, 1)
	require.Len(t, issuers[0].Denoms, 3)
}

func TestManageGasPrices1(t *testing.T) {
	ctx, keeper, _, _ := createTestComponents(t)

	var (
		accAuthority = mustParseAddress("emoney1kt0vh0ttget0xx77g6d3ttnvq2lnxx6vp3uyl0")
		accRandom    = mustParseAddress("emoney17up20gamd0vh6g9ne0uh67hx8xhyfrv2lyazgu")
	)

	keeper.SetAuthority(ctx, accAuthority)

	gasPrices := keeper.GetGasPrices(ctx)
	require.True(t, gasPrices.Empty())

	coins, _ := sdk.ParseDecCoins("0.0005eeur,0.000001echf")

	require.Panics(t, func() {
		keeper.SetGasPrices(ctx, accRandom, coins)
	})

	res, err := keeper.SetGasPrices(ctx, accAuthority, sdk.NewDecCoins())
	require.True(t, err == nil, res.Log)

	res, err = keeper.SetGasPrices(ctx, accAuthority, coins)
	require.True(t, err == nil, res.Log)

	gasPrices = keeper.GetGasPrices(ctx)
	require.Equal(t, coins, gasPrices)

	// Do not allow fees to be set in token denominations that are not present in the chain
	coins, _ = sdk.ParseDecCoins("0.0005eeur,0.000001echf,0.0000001esek")
	res, err = keeper.SetGasPrices(ctx, accAuthority, coins)
	require.True(t, types.ErrUnknownDenom.Is(err))
}

func TestManageGasPrices2(t *testing.T) {
	ctx, keeper, _, gpk := createTestComponents(t)

	// Manually write gas prices to appstate, circumventing the keeper
	setGasPrices := func(gp sdk.DecCoins) {
		bz := types.ModuleCdc.LegacyAmino.MustMarshalBinaryLengthPrefixed(gp)
		store := ctx.KVStore(keeper.storeKey)
		store.Set([]byte(keyGasPrices), bz)
	}

	gp, _ := sdk.ParseDecCoins("0.00005eeur")
	setGasPrices(gp)

	require.True(t, gpk.gasPrices.IsZero())
	BeginBlocker(ctx, keeper)
	require.Equal(t, gp, gpk.gasPrices)

	// Ensure that the initialization can only be invoked once

	gp2, _ := sdk.ParseDecCoins("0.003eusd")
	setGasPrices(gp2)

	BeginBlocker(ctx, keeper)
	// Verify that the gas prices remain the same
	require.Equal(t, gp, gpk.gasPrices)
}

func createTestComponents(t *testing.T) (sdk.Context, Keeper, issuer.Keeper, *mockGasPricesKeeper) {
	t.Helper()
	encConfig := MakeTestEncodingConfig()
	var (
		bankKey    = sdk.NewKVStoreKey(banktypes.ModuleName)
		authCapKey = sdk.NewKVStoreKey("authCapKey")
		keyParams  = sdk.NewKVStoreKey("params")
		authKey    = sdk.NewKVStoreKey(authtypes.StoreKey)
		tkeyParams = sdk.NewTransientStoreKey("transient_params")
		keyIssuer  = sdk.NewKVStoreKey(issuer.ModuleName)

		blockedAddr = make(map[string]bool)
	)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(authKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(bankKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyIssuer, sdk.StoreTypeIAVL, db)

	err := ms.LoadLatestVersion()
	require.NoError(t, err)

	maccPerms := map[string][]string{
		types.ModuleName: {authtypes.Minter},
	}

	ctx := sdk.NewContext(ms, tmproto.Header{ChainID: "test-chain"}, true, log.NewNopLogger())
	var (
		pk = paramskeeper.NewKeeper(encConfig.Marshaler, encConfig.Amino, keyParams, tkeyParams)
		ak = authkeeper.NewAccountKeeper(
			encConfig.Marshaler, authCapKey, pk.Subspace(authtypes.ModuleName), authtypes.ProtoBaseAccount, maccPerms,
		)
		bk = bankkeeper.NewBaseKeeper(
			encConfig.Marshaler, bankKey, ak, pk.Subspace(banktypes.ModuleName), blockedAddr,
		)
		lpk = liquidityprovider.NewKeeper(ak, bk)
		ik  = issuer.NewKeeper(keyIssuer, lpk, mockInflationKeeper{})
	)

	bk.SetSupply(ctx, banktypes.NewSupply(
		sdk.NewCoins(
			sdk.NewCoin("echf", sdk.NewInt(5000)),
			sdk.NewCoin("eeur", sdk.NewInt(5000)),
		)))

	gpk := new(mockGasPricesKeeper)
	keeper := NewKeeper(authKey, ik, bk, gpk)

	return ctx, keeper, ik, gpk
}

type mockGasPricesKeeper struct {
	gasPrices sdk.DecCoins
}

func (m *mockGasPricesKeeper) SetMinimumGasPrices(gasPricesStr string) error {
	if coins, err := sdk.ParseDecCoins(gasPricesStr); err != nil {
		return err
	} else {
		m.gasPrices = coins
	}

	return nil
}

type mockInflationKeeper struct{}

func (m mockInflationKeeper) SetInflation(ctx sdk.Context, inflation sdk.Dec, denom string) (_ *sdk.Result, _ error) {
	return
}

func (m mockInflationKeeper) AddDenoms(sdk.Context, []string) (_ *sdk.Result, _ error) {
	return
}

func MakeTestEncodingConfig() simappparams.EncodingConfig {
	cdc := codec.NewLegacyAmino()
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)

	encodingConfig := simappparams.EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaler:         marshaler,
		TxConfig:          tx.NewTxConfig(marshaler, tx.DefaultSignModes),
		Amino:             cdc,
	}

	std.RegisterLegacyAminoCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	ModuleBasics := module.NewBasicManager(
		bank.AppModuleBasic{},
		auth.AppModuleBasic{},
	)

	ModuleBasics.RegisterLegacyAminoCodec(encodingConfig.Amino)
	ModuleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	return encodingConfig
}

func mustParseAddress(address string) sdk.AccAddress {
	a, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		panic(err)
	}
	return a
}
