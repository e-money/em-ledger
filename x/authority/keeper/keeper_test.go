package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"testing"

	apptypes "emoney/types"
	"emoney/x/authority/types"
	"emoney/x/issuer"
	"emoney/x/liquidityprovider"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
)

func init() {
	// Be able to parse emoney bech32 encoded addresses.
	apptypes.ConfigureSDK()
}

func TestDenoms(t *testing.T) {
	require.True(t, validateDenom("x2eur"))
	require.False(t, validateDenom("X2EUR"))
	require.False(t, validateDenom("123456"))
}

func TestAuthorityBasicPersistence(t *testing.T) {
	ctx, keeper, _ := createTestComponents(t)

	require.Panics(t, func() {
		// Keeper must panic if no authority has been specified
		keeper.getAuthority(ctx)
	})

	acc, _ := sdk.AccAddressFromBech32("emoney1kt0vh0ttget0xx77g6d3ttnvq2lnxx6vp3uyl0")
	keeper.setAuthority(ctx, acc)

	authority := keeper.getAuthority(ctx)
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

	keeper.setAuthority(ctx, accAuthority)
	keeper.MustBeAuthority(ctx, accAuthority)

	require.Panics(t, func() {
		keeper.MustBeAuthority(ctx, acc2)
	})

	// Authority can only be specified once, preferably during genesis
	require.Panics(t, func() {
		keeper.setAuthority(ctx, acc2)
	})
}

func TestCreateAndRevokeIssuer(t *testing.T) {
	ctx, keeper, ik := createTestComponents(t)

	var (
		accAuthority = mustParseAddress("emoney1kt0vh0ttget0xx77g6d3ttnvq2lnxx6vp3uyl0")
		issuer1      = mustParseAddress("emoney17up20gamd0vh6g9ne0uh67hx8xhyfrv2lyazgu")
		issuer2      = mustParseAddress("emoney1dgkjvr2kkrp0xc5qn66g23us779q2dmgle5aum")
	)

	keeper.setAuthority(ctx, accAuthority)

	err := keeper.CreateIssuer(ctx, accAuthority, issuer1, []string{"x2eur", "x0jpy"})
	require.Nil(t, err)

	err = keeper.CreateIssuer(ctx, accAuthority, issuer2, []string{"x2chf", "x2gbp", "x2eur"})
	require.NotNil(t, err) // Must fail due to duplicate token denomination

	err = keeper.CreateIssuer(ctx, accAuthority, issuer2, []string{"x2chf", "x2gbp"})
	require.Nil(t, err)
	require.Len(t, ik.GetIssuers(ctx), 2)

	err = keeper.DestroyIssuer(ctx, accAuthority, issuer2)
	require.Nil(t, err)
	require.Len(t, ik.GetIssuers(ctx), 1)

	require.Panics(t, func() {
		// Make sure only authority key can destroy an issuer
		keeper.DestroyIssuer(ctx, issuer1, issuer2)
	})

	err = keeper.DestroyIssuer(ctx, accAuthority, issuer2)
	require.NotNil(t, err)
	require.Len(t, ik.GetIssuers(ctx), 1)

	err = keeper.DestroyIssuer(ctx, accAuthority, issuer1)
	require.Nil(t, err)
	require.Empty(t, ik.GetIssuers(ctx))
}

func createTestComponents(t *testing.T) (sdk.Context, *Keeper, issuer.Keeper) {
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
		bk  = bank.NewBaseKeeper(ak, pk.Subspace(bank.DefaultParamspace), bank.DefaultCodespace)
		sk  = supply.NewKeeper(cdc, keySupply, ak, bk, supply.DefaultCodespace, maccPerms)
		lpk = liquidityprovider.NewKeeper(ak, sk)
		ik  = issuer.NewKeeper(keySupply, lpk)
	)

	// Empty supply
	sk.SetSupply(ctx, supply.NewSupply(sdk.NewCoins()))

	keeper := NewKeeper(keyAuthority, ik)

	return ctx, keeper, ik
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
	types.RegisterCodec(cdc)

	return
}

func mustParseAddress(address string) sdk.AccAddress {
	a, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		panic(err)
	}
	return a
}
