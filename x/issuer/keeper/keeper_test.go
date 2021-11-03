// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package keeper

import (
	"sort"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	emauthtypes "github.com/e-money/em-ledger/x/authority/types"
	"github.com/e-money/em-ledger/x/issuer/types"
	"github.com/e-money/em-ledger/x/liquidityprovider"
	lptypes "github.com/e-money/em-ledger/x/liquidityprovider/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

func init() {
	// Be able to parse emoney bech32 encoded addresses.
	apptypes.ConfigureSDK()
}

func getDenomsMetadata(denoms []string) []emauthtypes.Denomination {
	md := make([]emauthtypes.Denomination, len(denoms))
	for i, denom := range denoms {
		md[i].Base = denom
	}
	return md
}

func TestAddIssuer(t *testing.T) {
	ctx, _, _, keeper, _ := createTestComponents(t)

	var (
		acc1, _      = sdk.AccAddressFromBech32("emoney1kt0vh0ttget0xx77g6d3ttnvq2lnxx6vp3uyl0")
		acc2, _      = sdk.AccAddressFromBech32("emoney17up20gamd0vh6g9ne0uh67hx8xhyfrv2lyazgu")
		randomacc, _ = sdk.AccAddressFromBech32("emoney1dgkjvr2kkrp0xc5qn66g23us779q2dmgle5aum")
		issuer1      = types.NewIssuer(acc1, "eeur", "ejpy")
		issuer2      = types.NewIssuer(acc2, "echf")
	)

	require.True(t, issuer1.IsValid())
	require.True(t, issuer2.IsValid())

	_, err := keeper.AddIssuer(ctx, issuer1, getDenomsMetadata(issuer1.Denoms))
	require.NoError(t, err)
	_, err = keeper.AddIssuer(ctx, issuer1, getDenomsMetadata(issuer1.Denoms))
	require.Error(t, err)
	_, err = keeper.AddIssuer(ctx, types.NewIssuer(acc1, "edkk"), []emauthtypes.Denomination{{Base: "edkk"}})
	require.NoError(t, err)

	require.Len(t, keeper.GetIssuers(ctx), 1)

	keeper.AddIssuer(ctx, issuer2, getDenomsMetadata(issuer2.Denoms))
	require.Len(t, keeper.GetIssuers(ctx), 2)
	require.Len(t, collectDenoms(keeper.GetIssuers(ctx)), 4)

	issuer, _ := keeper.mustBeIssuer(ctx, acc2.String())
	require.Equal(t, issuer2, issuer)

	_, err = keeper.mustBeIssuer(ctx, randomacc.String())
	require.Error(t, err)

	_, err = keeper.mustBeIssuer(ctx, "")
	require.Error(t, err)
}

func denomNotFound(ctx sdk.Context, t *testing.T, bk types.BankKeeper, denom string) {
	stateDenom := bk.GetDenomMetaData(ctx, denom)
	require.Empty(t, stateDenom.Base)
}

func denomFound(ctx sdk.Context, t *testing.T, bk types.BankKeeper, denom string) {
	stateDenom := bk.GetDenomMetaData(ctx, denom)
	require.NotEmpty(t, stateDenom.Base)
}

func TestAddDenomMetadata(t *testing.T) {
	ctx, _, _, keeper, bk := createTestComponents(t)

	var (
		acc1, _ = sdk.AccAddressFromBech32("emoney1kt0vh0ttget0xx77g6d3ttnvq2lnxx6vp3uyl0")
		issuer1 = types.NewIssuer(acc1, "eeur", "ejpy", "echf")
	)

	denomNotFound(ctx, t, bk, "eeur")
	denomNotFound(ctx, t, bk, "ejpy")
	denomNotFound(ctx, t, bk, "echf")

	_, err := keeper.AddIssuer(ctx, issuer1, getDenomsMetadata(issuer1.Denoms))
	require.NoError(t, err)
	denomFound(ctx, t, bk, "eeur")
	denomFound(ctx, t, bk, "ejpy")
	denomFound(ctx, t, bk, "echf")
	denomNotFound(ctx, t, bk, "enok")

	denomNotFound(ctx, t, bk, "edkk")
	_, err = keeper.AddIssuer(ctx, types.NewIssuer(acc1, "edkk"), []emauthtypes.Denomination{{Base: "edkk"}})
	require.NoError(t, err)
	denomFound(ctx, t, bk, "edkk")
}

func TestRemoveIssuer(t *testing.T) {
	ctx, _, _, keeper, _ := createTestComponents(t)

	acc1, _ := sdk.AccAddressFromBech32("emoney1kt0vh0ttget0xx77g6d3ttnvq2lnxx6vp3uyl0")
	acc2, _ := sdk.AccAddressFromBech32("emoney17up20gamd0vh6g9ne0uh67hx8xhyfrv2lyazgu")

	issuer := types.NewIssuer(acc1, "eeur", "ejpy")

	_, err := keeper.AddIssuer(ctx, issuer, getDenomsMetadata(issuer.Denoms))
	require.NoError(t, err)
	require.Len(t, keeper.GetIssuers(ctx), 1)

	_, err = keeper.RemoveIssuer(ctx, acc2)
	require.Error(t, err)
	require.Len(t, keeper.GetIssuers(ctx), 1)

	_, err = keeper.RemoveIssuer(ctx, acc1)
	require.NoError(t, err)
	require.Empty(t, keeper.GetIssuers(ctx))
}

func TestIssuerModifyLiquidityProvider(t *testing.T) {
	ctx, ak, lpk, keeper, _ := createTestComponents(t)

	var (
		iacc, _  = sdk.AccAddressFromBech32("emoney1kt0vh0ttget0xx77g6d3ttnvq2lnxx6vp3uyl0")
		lpacc, _ = sdk.AccAddressFromBech32("emoney17up20gamd0vh6g9ne0uh67hx8xhyfrv2lyazgu")
	)

	ak.SetAccount(ctx, ak.NewAccountWithAddress(ctx, lpacc))

	issuer := types.NewIssuer(iacc, "eeur", "ejpy")

	keeper.AddIssuer(ctx, issuer, getDenomsMetadata(issuer.Denoms))
	mintable := MustParseCoins("100000eeur,5000ejpy")

	_, err := keeper.IncreaseMintableAmountOfLiquidityProvider(ctx, lpacc, iacc, mintable)
	require.NoError(t, err)

	_, err = keeper.IncreaseMintableAmountOfLiquidityProvider(ctx, lpacc, iacc, mintable)
	require.NoError(t, err)

	// Verify the two increases in mintable balance
	a := lpk.GetLiquidityProviderAccount(ctx, lpacc)
	expected := MustParseCoins("200000eeur,10000ejpy")
	require.Equal(t, expected, a.Mintable)

	// Decrease the mintable amount too much
	mintable, _ = sdk.ParseCoinsNormalized("400000eeur")
	_, err = keeper.DecreaseMintableAmountOfLiquidityProvider(ctx, lpacc, iacc, mintable)
	require.NotNil(t, err)

	// Verify unchanged mintable amount
	require.Equal(t, expected, a.Mintable)

	// Decrease mintable balance.
	mintable = MustParseCoins("50000eeur, 2000ejpy")
	_, err = keeper.DecreaseMintableAmountOfLiquidityProvider(ctx, lpacc, iacc, mintable)
	require.NoError(t, err)

	expected = MustParseCoins("150000eeur,8000ejpy")
	a = lpk.GetLiquidityProviderAccount(ctx, lpacc)
	require.Equal(t, expected.String(), a.Mintable.String())
}

func TestAddAndRevokeLiquidityProvider(t *testing.T) {
	ctx, ak, _, keeper, _ := createTestComponents(t)

	var (
		iacc, _      = sdk.AccAddressFromBech32("emoney1kt0vh0ttget0xx77g6d3ttnvq2lnxx6vp3uyl0")
		lpacc, _     = sdk.AccAddressFromBech32("emoney17up20gamd0vh6g9ne0uh67hx8xhyfrv2lyazgu")
		randomacc, _ = sdk.AccAddressFromBech32("emoney1dgkjvr2kkrp0xc5qn66g23us779q2dmgle5aum")
	)

	ak.SetAccount(ctx, ak.NewAccountWithAddress(ctx, lpacc))

	keeper.AddIssuer(ctx, types.NewIssuer(iacc, "eeur", "ejpy"), []emauthtypes.Denomination{{Base: "eeur"}, {Base: "ejpy"}})

	mintable := MustParseCoins("100000eeur,5000ejpy")

	// Ensure that a random account can't create a LP
	_, err := keeper.IncreaseMintableAmountOfLiquidityProvider(ctx, lpacc, randomacc, mintable)
	require.Error(t, err)

	_, err = keeper.IncreaseMintableAmountOfLiquidityProvider(ctx, lpacc, iacc, mintable)
	require.NoError(t, err)

	// Make sure a random account can't revoke LP status
	_, err = keeper.RevokeLiquidityProvider(ctx, lpacc, randomacc)
	require.Error(t, err)

	_, err = keeper.RevokeLiquidityProvider(ctx, lpacc, iacc)
	require.NoError(t, err)
	require.IsType(t, &authtypes.BaseAccount{}, ak.GetAccount(ctx, lpacc))
}

func TestDoubleLiquidityProvider(t *testing.T) {
	// Two issuers provide lp status to the same account. Ensure revocation is isolated.
	ctx, ak, lpk, keeper, _ := createTestComponents(t)

	var (
		issuer1, _ = sdk.AccAddressFromBech32("emoney1kt0vh0ttget0xx77g6d3ttnvq2lnxx6vp3uyl0")
		issuer2, _ = sdk.AccAddressFromBech32("emoney17up20gamd0vh6g9ne0uh67hx8xhyfrv2lyazgu")
		lp, _      = sdk.AccAddressFromBech32("emoney1dgkjvr2kkrp0xc5qn66g23us779q2dmgle5aum")
	)

	ak.SetAccount(ctx, ak.NewAccountWithAddress(ctx, lp))
	keeper.AddIssuer(ctx, types.NewIssuer(issuer1, "eeur", "ejpy"), []emauthtypes.Denomination{{Base: "eeur"}, {Base: "ejpy"}})
	keeper.AddIssuer(ctx, types.NewIssuer(issuer2, "edkk", "esek"), []emauthtypes.Denomination{{Base: "edkk"}, {Base: "esek"}})

	mintable1 := MustParseCoins("100000eeur,5000ejpy")
	mintable2 := MustParseCoins("250000edkk,1000esek")

	_, err := keeper.IncreaseMintableAmountOfLiquidityProvider(ctx, lp, issuer1, mintable1)
	require.NoError(t, err)

	// Attempt to revoke liquidity given by other issuer
	_, err = keeper.RevokeLiquidityProvider(ctx, lp, issuer2)
	require.Error(t, err)

	_, err = keeper.IncreaseMintableAmountOfLiquidityProvider(ctx, lp, issuer2, mintable2)
	require.NoError(t, err)

	lpAccount := lpk.GetLiquidityProviderAccount(ctx, lp)
	require.Len(t, lpAccount.Mintable, 4)

	_, err = keeper.RevokeLiquidityProvider(ctx, lp, issuer1)
	require.NoError(t, err)

	lpAccount = lpk.GetLiquidityProviderAccount(ctx, lp)
	require.Len(t, lpAccount.Mintable, 2)

	_, err = keeper.RevokeLiquidityProvider(ctx, lp, issuer2)
	require.NoError(t, err)
	require.IsType(t, &authtypes.BaseAccount{}, ak.GetAccount(ctx, lp))
}

func TestCollectDenominations(t *testing.T) {
	issuers := []types.Issuer{
		{
			Denoms: []string{"eeur", "ejpy"},
		},
		{
			Denoms: []string{"echf", "edkk"},
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

func createTestComponents(t *testing.T) (sdk.Context, authkeeper.AccountKeeper, liquidityprovider.Keeper, Keeper, types.BankKeeper) {
	return createTestComponentsWithEncodingConfig(t, MakeTestEncodingConfig())
}

func createTestComponentsWithEncodingConfig(t *testing.T, encConfig simappparams.EncodingConfig) (sdk.Context, authkeeper.AccountKeeper, liquidityprovider.Keeper, Keeper, types.BankKeeper) {
	t.Helper()
	var (
		bankKey    = sdk.NewKVStoreKey(banktypes.ModuleName)
		authCapKey = sdk.NewKVStoreKey("authCapKey")
		keyParams  = sdk.NewKVStoreKey("params")
		stakingKey = sdk.NewKVStoreKey("staking")
		authKey    = sdk.NewKVStoreKey(authtypes.StoreKey)
		tkeyParams = sdk.NewTransientStoreKey("transient_params")
		issuerKey  = sdk.NewKVStoreKey(types.StoreKey)
		lpKey      = sdk.NewKVStoreKey(lptypes.StoreKey)

		blockedAddrs = make(map[string]bool)
		maccPerms    = map[string][]string{
			types.ModuleName: {authtypes.Minter},
		}
	)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(stakingKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(authKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(bankKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(lpKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(issuerKey, sdk.StoreTypeIAVL, db)

	err := ms.LoadLatestVersion()
	require.NoError(t, err)

	ctx := sdk.NewContext(ms, tmproto.Header{ChainID: "supply-chain"}, true, log.NewNopLogger())

	var (
		pk = paramskeeper.NewKeeper(encConfig.Marshaler, encConfig.Amino, keyParams, tkeyParams)
		ak = authkeeper.NewAccountKeeper(
			encConfig.Marshaler, authCapKey, pk.Subspace(authtypes.ModuleName), authtypes.ProtoBaseAccount, maccPerms,
		)
		bk = bankkeeper.NewBaseKeeper(
			encConfig.Marshaler, bankKey, ak, pk.Subspace(banktypes.ModuleName), blockedAddrs,
		)
	)

	// Empty supply
	bk.SetSupply(ctx, banktypes.NewSupply(sdk.NewCoins()))

	lpk := liquidityprovider.NewKeeper(encConfig.Marshaler, lpKey, bk)

	keeper := NewKeeper(encConfig.Marshaler, issuerKey, lpk, mockInflationKeeper{}, bk)
	return ctx, ak, lpk, keeper, bk
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
		liquidityprovider.AppModuleBasic{},
	)

	ModuleBasics.RegisterLegacyAminoCodec(encodingConfig.Amino)
	ModuleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	return encodingConfig
}

func MustParseCoins(coins string) sdk.Coins {
	result, err := sdk.ParseCoinsNormalized(coins)
	if err != nil {
		panic(err)
	}

	return result
}
