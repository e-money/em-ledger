// This software is Copyright (c) 2019 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"testing"

	apptypes "github.com/e-money/em-ledger/types"
	"github.com/e-money/em-ledger/x/authority/types"
	"github.com/e-money/em-ledger/x/issuer"
	"github.com/e-money/em-ledger/x/liquidityprovider"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

func init() {
	// Be able to parse emoney bech32 encoded addresses.
	apptypes.ConfigureSDK()
}

func TestAuthorityBasicPersistence(t *testing.T) {
	ctx, keeper, _ := createTestComponents(t)

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
	ctx, keeper, _ := createTestComponents(t)

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
	ctx, keeper, ik := createTestComponents(t)

	var (
		accAuthority = mustParseAddress("emoney1kt0vh0ttget0xx77g6d3ttnvq2lnxx6vp3uyl0")
		issuer1      = mustParseAddress("emoney17up20gamd0vh6g9ne0uh67hx8xhyfrv2lyazgu")
		issuer2      = mustParseAddress("emoney1dgkjvr2kkrp0xc5qn66g23us779q2dmgle5aum")
	)

	keeper.SetAuthority(ctx, accAuthority)

	result := keeper.CreateIssuer(ctx, accAuthority, issuer1, []string{"eeur", "ejpy"})
	require.True(t, result.IsOK())

	result = keeper.CreateIssuer(ctx, accAuthority, issuer2, []string{"echf", "egbp", "eeur"})
	require.False(t, result.IsOK()) // Must fail due to duplicate token denomination

	result = keeper.CreateIssuer(ctx, accAuthority, issuer2, []string{"echf", "egbp"})
	require.True(t, result.IsOK())
	require.Len(t, ik.GetIssuers(ctx), 2)

	result = keeper.DestroyIssuer(ctx, accAuthority, issuer2)
	require.True(t, result.IsOK())
	require.Len(t, ik.GetIssuers(ctx), 1)

	require.Panics(t, func() {
		// Make sure only authority key can destroy an issuer
		keeper.DestroyIssuer(ctx, issuer1, issuer2)
	})

	result = keeper.DestroyIssuer(ctx, accAuthority, issuer2)
	require.False(t, result.IsOK())
	require.Len(t, ik.GetIssuers(ctx), 1)

	result = keeper.DestroyIssuer(ctx, accAuthority, issuer1)
	require.True(t, result.IsOK())
	require.Empty(t, ik.GetIssuers(ctx))
}

func TestAddMultipleDenomsSameIssuer(t *testing.T) {
	ctx, keeper, ik := createTestComponents(t)

	var (
		accAuthority = mustParseAddress("emoney1kt0vh0ttget0xx77g6d3ttnvq2lnxx6vp3uyl0")
		accIssuer    = mustParseAddress("emoney17up20gamd0vh6g9ne0uh67hx8xhyfrv2lyazgu")
	)

	keeper.SetAuthority(ctx, accAuthority)

	result := keeper.CreateIssuer(ctx, accAuthority, accIssuer, []string{"eeur", "ejpy"})
	require.True(t, result.IsOK())

	result = keeper.CreateIssuer(ctx, accAuthority, accIssuer, []string{"ekrw"})
	require.True(t, result.IsOK())
	issuers := ik.GetIssuers(ctx)

	// Ensure that the denomination has been added to the existing issuer, not to a new entry with the same key
	require.Len(t, issuers, 1)
	require.Len(t, issuers[0].Denoms, 3)
}

func TestManageGasPrices(t *testing.T) {
	ctx, keeper, _ := createTestComponents(t)

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

	res := keeper.SetGasPrices(ctx, accAuthority, coins)
	require.True(t, res.IsOK(), res.Log)

	gasPrices = keeper.GetGasPrices(ctx)
	require.Equal(t, coins, gasPrices)
}

func createTestComponents(t *testing.T) (sdk.Context, Keeper, issuer.Keeper) {
	cdc := makeTestCodec()

	logger := log.NewNopLogger() // Default
	//logger = log.NewTMLogger(os.Stdout) // Override to see output

	var (
		keyAuthority = sdk.NewKVStoreKey(types.ModuleName)
		keyAcc       = sdk.NewKVStoreKey(auth.StoreKey)
		keyParams    = sdk.NewKVStoreKey(params.StoreKey)
		keySupply    = sdk.NewKVStoreKey(supply.StoreKey)
		keyIssuer    = sdk.NewKVStoreKey(issuer.ModuleName)
		tkeyParams   = sdk.NewTransientStoreKey(params.TStoreKey)
	)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(keyAuthority, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keySupply, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyIssuer, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)

	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "supply-chain"}, true, logger)

	maccPerms := map[string][]string{
		types.ModuleName: {supply.Minter},
	}

	var (
		pk  = params.NewKeeper(cdc, keyParams, tkeyParams, params.DefaultCodespace)
		ak  = auth.NewAccountKeeper(cdc, keyAcc, pk.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount)
		bk  = bank.NewBaseKeeper(ak, pk.Subspace(bank.DefaultParamspace), bank.DefaultCodespace, make(map[string]bool))
		sk  = supply.NewKeeper(cdc, keySupply, ak, bk, maccPerms)
		lpk = liquidityprovider.NewKeeper(ak, sk)
		ik  = issuer.NewKeeper(keySupply, lpk, mockInflationKeeper{})
	)

	// Empty supply
	sk.SetSupply(ctx, supply.NewSupply(sdk.NewCoins()))

	keeper := NewKeeper(keyAuthority, ik, mockGasPricesKeeper{})

	return ctx, keeper, ik
}

type mockGasPricesKeeper struct {
	gasPrices sdk.DecCoins
}

func (m mockGasPricesKeeper) SetMinimumGasPrices(gasPricesStr string) error {
	if coins, err := sdk.ParseDecCoins(gasPricesStr); err != nil {
		return err
	} else {
		m.gasPrices = coins
	}

	return nil
}

type mockInflationKeeper struct{}

func (m mockInflationKeeper) SetInflation(ctx sdk.Context, inflation sdk.Dec, denom string) (_ sdk.Result) {
	return
}

func (m mockInflationKeeper) AddDenoms(sdk.Context, []string) (_ sdk.Result) {
	return
}

func makeTestCodec() (cdc *codec.Codec) {
	cdc = codec.New()

	bank.RegisterCodec(cdc)
	auth.RegisterCodec(cdc)
	types.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	supply.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	liquidityprovider.RegisterCodec(cdc)

	return
}

func mustParseAddress(address string) sdk.AccAddress {
	a, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		panic(err)
	}
	return a
}
