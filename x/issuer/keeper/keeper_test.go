// This software is Copyright (c) 2019 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package keeper

import (
	apptypes "github.com/e-money/em-ledger/types"
	"github.com/e-money/em-ledger/x/liquidityprovider"
	"sort"
	"testing"

	"github.com/e-money/em-ledger/x/issuer/types"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/supply"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

func init() {
	// Be able to parse emoney bech32 encoded addresses.
	apptypes.ConfigureSDK()
}

func TestAddIssuer(t *testing.T) {
	ctx, _, _, keeper := createTestComponents(t)

	var (
		acc1, _      = sdk.AccAddressFromBech32("emoney1kt0vh0ttget0xx77g6d3ttnvq2lnxx6vp3uyl0")
		acc2, _      = sdk.AccAddressFromBech32("emoney17up20gamd0vh6g9ne0uh67hx8xhyfrv2lyazgu")
		randomacc, _ = sdk.AccAddressFromBech32("emoney1dgkjvr2kkrp0xc5qn66g23us779q2dmgle5aum")
		issuer1      = types.NewIssuer(acc1, "eeur", "ejpy")
		issuer2      = types.NewIssuer(acc2, "echf")
	)

	require.True(t, issuer1.IsValid())
	require.True(t, issuer2.IsValid())

	result := keeper.AddIssuer(ctx, issuer1)
	require.True(t, result.IsOK())
	result = keeper.AddIssuer(ctx, issuer1)
	require.False(t, result.IsOK())
	result = keeper.AddIssuer(ctx, types.NewIssuer(acc1, "edkk"))
	require.True(t, result.IsOK())

	require.Len(t, keeper.GetIssuers(ctx), 1)

	keeper.AddIssuer(ctx, issuer2)
	require.Len(t, keeper.GetIssuers(ctx), 2)
	require.Len(t, collectDenoms(keeper.GetIssuers(ctx)), 4)

	issuer, _ := keeper.mustBeIssuer(ctx, acc2)
	require.Equal(t, issuer2, issuer)

	_, err := keeper.mustBeIssuer(ctx, randomacc)
	require.Error(t, err)

	_, err = keeper.mustBeIssuer(ctx, nil)
	require.Error(t, err)
}

func TestRemoveIssuer(t *testing.T) {
	ctx, _, _, keeper := createTestComponents(t)

	acc1, _ := sdk.AccAddressFromBech32("emoney1kt0vh0ttget0xx77g6d3ttnvq2lnxx6vp3uyl0")
	acc2, _ := sdk.AccAddressFromBech32("emoney17up20gamd0vh6g9ne0uh67hx8xhyfrv2lyazgu")

	issuer := types.NewIssuer(acc1, "eeur", "ejpy")

	result := keeper.AddIssuer(ctx, issuer)
	require.True(t, result.IsOK())
	require.Len(t, keeper.GetIssuers(ctx), 1)

	result = keeper.RemoveIssuer(ctx, acc2)
	require.False(t, result.IsOK())
	require.Len(t, keeper.GetIssuers(ctx), 1)

	result = keeper.RemoveIssuer(ctx, acc1)
	require.True(t, result.IsOK())
	require.Empty(t, keeper.GetIssuers(ctx))
}

func TestIssuerModifyLiquidityProvider(t *testing.T) {
	ctx, ak, _, keeper := createTestComponents(t)

	var (
		iacc, _  = sdk.AccAddressFromBech32("emoney1kt0vh0ttget0xx77g6d3ttnvq2lnxx6vp3uyl0")
		lpacc, _ = sdk.AccAddressFromBech32("emoney17up20gamd0vh6g9ne0uh67hx8xhyfrv2lyazgu")
	)

	ak.SetAccount(ctx, ak.NewAccountWithAddress(ctx, lpacc))

	issuer := types.NewIssuer(iacc, "eeur", "ejpy")

	keeper.AddIssuer(ctx, issuer)
	mintable := MustParseCoins("100000eeur,5000ejpy")

	keeper.IncreaseMintableAmountOfLiquidityProvider(ctx, lpacc, issuer.Address, mintable)
	require.IsType(t, &liquidityprovider.Account{}, ak.GetAccount(ctx, lpacc))

	keeper.IncreaseMintableAmountOfLiquidityProvider(ctx, lpacc, issuer.Address, mintable)

	// Verify the two increases in mintable balance
	a := ak.GetAccount(ctx, lpacc).(*liquidityprovider.Account)
	expected := MustParseCoins("200000eeur,10000ejpy")
	require.Equal(t, expected, a.Mintable)

	// Decrease the mintable amount too much
	mintable, _ = sdk.ParseCoins("400000eeur")
	result := keeper.DecreaseMintableAmountOfLiquidityProvider(ctx, lpacc, issuer.Address, mintable)
	require.NotNil(t, result)

	// Verify unchanged mintable amount
	a = ak.GetAccount(ctx, lpacc).(*liquidityprovider.Account)
	require.Equal(t, expected, a.Mintable)

	// Decrease mintable balance.
	mintable = MustParseCoins("50000eeur, 2000ejpy")
	result = keeper.DecreaseMintableAmountOfLiquidityProvider(ctx, lpacc, issuer.Address, mintable)
	require.True(t, result.IsOK())

	expected = MustParseCoins("150000eeur,8000ejpy")
	a = ak.GetAccount(ctx, lpacc).(*liquidityprovider.Account)
	require.Equal(t, expected, a.Mintable)
}

func TestAddAndRevokeLiquidityProvider(t *testing.T) {
	ctx, ak, _, keeper := createTestComponents(t)

	var (
		iacc, _      = sdk.AccAddressFromBech32("emoney1kt0vh0ttget0xx77g6d3ttnvq2lnxx6vp3uyl0")
		lpacc, _     = sdk.AccAddressFromBech32("emoney17up20gamd0vh6g9ne0uh67hx8xhyfrv2lyazgu")
		randomacc, _ = sdk.AccAddressFromBech32("emoney1dgkjvr2kkrp0xc5qn66g23us779q2dmgle5aum")
	)

	ak.SetAccount(ctx, ak.NewAccountWithAddress(ctx, lpacc))

	keeper.AddIssuer(ctx, types.NewIssuer(iacc, "eeur", "ejpy"))

	mintable := MustParseCoins("100000eeur,5000ejpy")

	// Ensure that a random account can't create a LP
	res := keeper.IncreaseMintableAmountOfLiquidityProvider(ctx, lpacc, randomacc, mintable)
	require.False(t, res.IsOK())

	keeper.IncreaseMintableAmountOfLiquidityProvider(ctx, lpacc, iacc, mintable)
	require.IsType(t, &liquidityprovider.Account{}, ak.GetAccount(ctx, lpacc))

	// Make sure a random account can't revoke LP status
	res = keeper.RevokeLiquidityProvider(ctx, lpacc, randomacc)
	require.False(t, res.IsOK())

	result := keeper.RevokeLiquidityProvider(ctx, lpacc, iacc)
	require.True(t, result.IsOK(), "%v", result)
	require.IsType(t, &auth.BaseAccount{}, ak.GetAccount(ctx, lpacc))
}

func TestDoubleLiquidityProvider(t *testing.T) {
	// Two issuers provide lp status to the same account. Ensure revocation is isolated.
	ctx, ak, lpk, keeper := createTestComponents(t)

	var (
		issuer1, _ = sdk.AccAddressFromBech32("emoney1kt0vh0ttget0xx77g6d3ttnvq2lnxx6vp3uyl0")
		issuer2, _ = sdk.AccAddressFromBech32("emoney17up20gamd0vh6g9ne0uh67hx8xhyfrv2lyazgu")
		lp, _      = sdk.AccAddressFromBech32("emoney1dgkjvr2kkrp0xc5qn66g23us779q2dmgle5aum")
	)

	ak.SetAccount(ctx, ak.NewAccountWithAddress(ctx, lp))
	keeper.AddIssuer(ctx, types.NewIssuer(issuer1, "eeur", "ejpy"))
	keeper.AddIssuer(ctx, types.NewIssuer(issuer2, "edkk", "esek"))

	mintable1 := MustParseCoins("100000eeur,5000ejpy")
	mintable2 := MustParseCoins("250000edkk,1000esek")

	keeper.IncreaseMintableAmountOfLiquidityProvider(ctx, lp, issuer1, mintable1)

	// Attempt to revoke liquidity given by other issuer
	res := keeper.RevokeLiquidityProvider(ctx, lp, issuer2)
	require.False(t, res.IsOK())

	keeper.IncreaseMintableAmountOfLiquidityProvider(ctx, lp, issuer2, mintable2)

	lpAccount := lpk.GetLiquidityProviderAccount(ctx, lp)
	require.Len(t, lpAccount.Mintable, 4)

	res = keeper.RevokeLiquidityProvider(ctx, lp, issuer1)
	require.True(t, res.IsOK())

	lpAccount = lpk.GetLiquidityProviderAccount(ctx, lp)
	require.Len(t, lpAccount.Mintable, 2)

	res = keeper.RevokeLiquidityProvider(ctx, lp, issuer2)
	require.True(t, res.IsOK())
	require.IsType(t, &auth.BaseAccount{}, ak.GetAccount(ctx, lp))
}

func TestCollectDenominations(t *testing.T) {
	issuers := []types.Issuer{
		{
			Address: nil,
			Denoms:  []string{"eeur", "ejpy"},
		},
		{
			Address: nil,
			Denoms:  []string{"echf", "edkk"},
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

func TestRemoveDenom(t *testing.T) {
	coins := sdk.NewCoins(
		sdk.NewCoin("eeur", sdk.NewInt(5)),
		sdk.NewCoin("edkk", sdk.NewInt(5)),
		sdk.NewCoin("ejpy", sdk.NewInt(5)),
	)

	res := removeDenom(coins, "echf")
	require.EqualValues(t, coins, res)

	res = removeDenom(coins, "ejpy")
	require.Len(t, res, 2)
	require.EqualValues(t, coins[:len(coins)-1], res)
}

func createTestComponents(t *testing.T) (sdk.Context, auth.AccountKeeper, liquidityprovider.Keeper, Keeper) {
	cdc := makeTestCodec()

	var (
		keyAcc     = sdk.NewKVStoreKey(auth.StoreKey)
		keyParams  = sdk.NewKVStoreKey(params.StoreKey)
		keySupply  = sdk.NewKVStoreKey(supply.StoreKey)
		keyIssuer  = sdk.NewKVStoreKey(types.StoreKey)
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
	bk := bank.NewBaseKeeper(ak, pk.Subspace(bank.DefaultParamspace), bank.DefaultCodespace, make(map[string]bool))
	sk := supply.NewKeeper(cdc, keySupply, ak, bk, maccPerms)

	// Empty supply
	sk.SetSupply(ctx, supply.NewSupply(sdk.NewCoins()))

	lpk := liquidityprovider.NewKeeper(ak, sk)

	keeper := NewKeeper(keySupply, lpk, mockInflationKeeper{})
	return ctx, ak, lpk, keeper
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

func MustParseCoins(coins string) sdk.Coins {
	result, err := sdk.ParseCoins(coins)
	if err != nil {
		panic(err)
	}

	return result
}
