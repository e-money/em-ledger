package keeper

import (
	apptypes "emoney/types"
	"emoney/x/liquidityprovider"
	"sort"
	"testing"

	"emoney/x/issuer/types"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/supply"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
)

func init() {
	// Be able to parse emoney bech32 encoded addresses.
	apptypes.ConfigureSDK()
}

func TestAddIssuer(t *testing.T) {
	ctx, _, _, keeper := createTestComponents(t)

	acc1, _ := sdk.AccAddressFromBech32("emoney1kt0vh0ttget0xx77g6d3ttnvq2lnxx6vp3uyl0")

	var (
		issuer1 = types.NewIssuer(acc1, "x2eur", "x0jpy")
		issuer2 = types.NewIssuer(acc1, "x2chf")
	)

	require.True(t, issuer1.IsValid())
	require.True(t, issuer2.IsValid())

	err := keeper.AddIssuer(ctx, issuer1)
	require.Nil(t, err)
	err = keeper.AddIssuer(ctx, issuer1)
	require.NotNil(t, err)

	require.Len(t, keeper.GetIssuers(ctx), 1)

	keeper.AddIssuer(ctx, issuer2)
	require.Len(t, keeper.GetIssuers(ctx), 2)
	require.Len(t, collectDenoms(keeper.GetIssuers(ctx)), 3)

	issuer := keeper.mustBeIssuer(ctx, acc1)
	require.Equal(t, issuer1, issuer)

	acc2, _ := sdk.AccAddressFromBech32("emoney17up20gamd0vh6g9ne0uh67hx8xhyfrv2lyazgu")
	require.Panics(t, func() {
		// Function must panic if provided with a non-issuer account
		keeper.mustBeIssuer(ctx, acc2)
	})
	require.Panics(t, func() {
		// Function must panic if somehow provided with a nil address
		keeper.mustBeIssuer(ctx, nil)
	})
}

func TestRemoveIssuer(t *testing.T) {
	ctx, _, _, keeper := createTestComponents(t)

	acc1, _ := sdk.AccAddressFromBech32("emoney1kt0vh0ttget0xx77g6d3ttnvq2lnxx6vp3uyl0")
	acc2, _ := sdk.AccAddressFromBech32("emoney17up20gamd0vh6g9ne0uh67hx8xhyfrv2lyazgu")

	issuer := types.NewIssuer(acc1, "x2eur", "x0jpy")

	err := keeper.AddIssuer(ctx, issuer)
	require.Nil(t, err)
	require.Len(t, keeper.GetIssuers(ctx), 1)

	err = keeper.RemoveIssuer(ctx, acc2)
	require.NotNil(t, err)
	require.Len(t, keeper.GetIssuers(ctx), 1)

	err = keeper.RemoveIssuer(ctx, acc1)
	require.Nil(t, err)
	require.Empty(t, keeper.GetIssuers(ctx))
}

func TestIssuerModifyLiquidityProvider(t *testing.T) {
	ctx, ak, _, keeper := createTestComponents(t)

	var (
		iacc, _  = sdk.AccAddressFromBech32("emoney1kt0vh0ttget0xx77g6d3ttnvq2lnxx6vp3uyl0")
		lpacc, _ = sdk.AccAddressFromBech32("emoney17up20gamd0vh6g9ne0uh67hx8xhyfrv2lyazgu")
	)

	ak.SetAccount(ctx, ak.NewAccountWithAddress(ctx, lpacc))

	issuer := types.NewIssuer(iacc, "x2eur", "x0jpy")

	keeper.AddIssuer(ctx, issuer)
	credit := MustParseCoins("100000x2eur,5000x0jpy")

	keeper.IncreaseCreditOfLiquidityProvider(ctx, lpacc, issuer.Address, credit)
	require.IsType(t, &liquidityprovider.Account{}, ak.GetAccount(ctx, lpacc))

	keeper.IncreaseCreditOfLiquidityProvider(ctx, lpacc, issuer.Address, credit)

	// Verify the two increases in credit
	a := ak.GetAccount(ctx, lpacc).(*liquidityprovider.Account)
	expected := MustParseCoins("200000x2eur,10000x0jpy")
	require.Equal(t, expected, a.Credit)

	// Decrease the credit too much
	credit, _ = sdk.ParseCoins("400000x2eur")
	result := keeper.DecreaseCreditOfLiquidityProvider(ctx, lpacc, issuer.Address, credit)
	require.NotNil(t, result)

	// Verify unchanged credit
	a = ak.GetAccount(ctx, lpacc).(*liquidityprovider.Account)
	require.Equal(t, expected, a.Credit)

	// Decrease credit.
	credit = MustParseCoins("50000x2eur, 2000x0jpy")
	result = keeper.DecreaseCreditOfLiquidityProvider(ctx, lpacc, issuer.Address, credit)
	require.Nil(t, result)

	expected = MustParseCoins("150000x2eur,8000x0jpy")
	a = ak.GetAccount(ctx, lpacc).(*liquidityprovider.Account)
	require.Equal(t, expected, a.Credit)
}

func TestAddAndRevokeLiquidityProvider(t *testing.T) {
	ctx, ak, _, keeper := createTestComponents(t)

	var (
		iacc, _      = sdk.AccAddressFromBech32("emoney1kt0vh0ttget0xx77g6d3ttnvq2lnxx6vp3uyl0")
		lpacc, _     = sdk.AccAddressFromBech32("emoney17up20gamd0vh6g9ne0uh67hx8xhyfrv2lyazgu")
		randomacc, _ = sdk.AccAddressFromBech32("emoney17up20gamd0vh6g9ne0uh67hx8xhyfrv2lyazgu")
	)

	ak.SetAccount(ctx, ak.NewAccountWithAddress(ctx, lpacc))

	keeper.AddIssuer(ctx, types.NewIssuer(iacc, "x2eur", "x0jpy"))

	credit := MustParseCoins("100000x2eur,5000x0jpy")

	// Ensure that a random account can't create a LP
	require.Panics(t, func() {
		keeper.IncreaseCreditOfLiquidityProvider(ctx, lpacc, randomacc, credit)
	})

	keeper.IncreaseCreditOfLiquidityProvider(ctx, lpacc, iacc, credit)
	require.IsType(t, &liquidityprovider.Account{}, ak.GetAccount(ctx, lpacc))

	// Make sure a random account can't revoke LP status
	require.Panics(t, func() {
		keeper.RevokeLiquidityProvider(ctx, lpacc, randomacc)
	})

	err := keeper.RevokeLiquidityProvider(ctx, lpacc, iacc)
	require.Nil(t, err, "%v", err)
	require.IsType(t, &auth.BaseAccount{}, ak.GetAccount(ctx, lpacc))
}

func TestCollectDenominations(t *testing.T) {
	issuers := []types.Issuer{
		{
			Address: nil,
			Denoms:  []string{"x2eur", "x0jpy"},
		},
		{
			Address: nil,
			Denoms:  []string{"x2chf", "x0dkk"},
		},
	}

	denoms := collectDenoms(issuers)
	require.Len(t, denoms, 4)
	require.True(t, sort.StringsAreSorted(denoms))
}

func TestAnyContains(t *testing.T) {
	// Test this basic plumbing, just to be sure
	input := []string{"bird", "apple", "ocean", "fork", "anchor"}
	sort.Strings(input)

	require.True(t, anyContained(input, "ocean", "flow"))
	require.True(t, anyContained(input, "anchor"))
	require.True(t, anyContained([]string{"bird"}, "bird"))

	require.False(t, anyContained(input, "flow", "eagle"))
	require.False(t, anyContained(input))
	require.False(t, anyContained(make([]string, 0), "ocean"))
}

func createTestComponents(t *testing.T) (sdk.Context, auth.AccountKeeper, liquidityprovider.Keeper, Keeper) {
	cdc := makeTestCodec()

	var (
		keyAcc     = sdk.NewKVStoreKey(auth.StoreKey)
		keyParams  = sdk.NewKVStoreKey(params.StoreKey)
		keySupply  = sdk.NewKVStoreKey(supply.StoreKey)
		keyIssuer  = sdk.NewKVStoreKey(types.ModuleName)
		tkeyParams = sdk.NewTransientStoreKey(params.TStoreKey)
	)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keySupply, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyIssuer, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	logger := log.NewNopLogger() // Default
	//logger = log.NewTMLogger(os.Stdout) // Override to see output

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "supply-chain"}, true, logger)

	maccPerms := map[string][]string{
		types.ModuleName: {supply.Minter},
	}

	pk := params.NewKeeper(cdc, keyParams, tkeyParams, params.DefaultCodespace)
	ak := auth.NewAccountKeeper(cdc, keyAcc, pk.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount)
	bk := bank.NewBaseKeeper(ak, pk.Subspace(bank.DefaultParamspace), bank.DefaultCodespace)
	sk := supply.NewKeeper(cdc, keySupply, ak, bk, supply.DefaultCodespace, maccPerms)

	// Empty supply
	sk.SetSupply(ctx, supply.NewSupply(sdk.NewCoins()))

	lpk := liquidityprovider.NewKeeper(ak, sk)

	keeper := NewKeeper(keySupply, lpk)

	return ctx, ak, lpk, keeper
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

func MustParseCoins(coins string) sdk.Coins {
	result, err := sdk.ParseCoins(coins)
	if err != nil {
		panic(err)
	}

	return result
}
