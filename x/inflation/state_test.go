// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package inflation

import (
	"testing"
	"time"

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
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/e-money/em-ledger/x/inflation/internal/keeper"
	"github.com/e-money/em-ledger/x/inflation/internal/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

func TestModule1(t *testing.T) {
	ctx, keeper, bankKeeper, _ := createTestComponents()

	initialEurAmount, _ := sdk.ParseCoinNormalized("1000000000eur")

	bankKeeper.SetSupply(ctx, banktypes.NewSupply(sdk.Coins{initialEurAmount}))

	currentTime := time.Now()
	ctx = ctx.WithBlockTime(currentTime).WithBlockHeight(55)
	BeginBlocker(ctx, keeper)

	keeper.AddDenoms(ctx, []string{"eur"})
	keeper.SetInflation(ctx, sdk.NewDecWithPrec(1, 2), "eur")

	for i := int64(1); i <= 10; i++ {
		currentTime = currentTime.Add(time.Minute)
		ctx = ctx.WithBlockTime(currentTime).WithBlockHeight(60 + 5*i)
		BeginBlocker(ctx, keeper)

		// Inflation of 1% per year on EUR1.000.000.000, each minute should add approximately 19 euro in interest.
		total := bankKeeper.GetSupply(ctx).GetTotal()
		minted := total.AmountOf("eur").Sub(initialEurAmount.Amount)
		require.True(t, minted.LT(sdk.NewInt(20).MulRaw(i)))
	}
}

// Verify that the newly minted tokens are sent to the correct modules
func TestModuleDestinations(t *testing.T) {
	ctx, keeper, bankKeeper, accountKeeper := createTestComponents()

	bankKeeper.SetSupply(ctx, banktypes.NewSupply(coins("400000000eur,400000000chf,100000000ungm")))

	currentTime := time.Now()
	ctx = ctx.WithBlockTime(currentTime).WithBlockHeight(55)

	BeginBlocker(ctx, keeper)

	keeper.AddDenoms(ctx, []string{"eur", "chf", "ungm"})
	keeper.SetInflation(ctx, sdk.NewDecWithPrec(1, 2), "eur")
	keeper.SetInflation(ctx, sdk.NewDecWithPrec(1, 2), "chf")
	keeper.SetInflation(ctx, sdk.NewDecWithPrec(10, 2), "ungm")

	for i := int64(1); i <= 10; i++ {
		currentTime = currentTime.Add(time.Minute)
		ctx = ctx.WithBlockTime(currentTime).WithBlockHeight(60 + 5*i)
		BeginBlocker(ctx, keeper)
	}

	// Stablecoin tokens should be in the buyback module's account
	buybackacc := accountKeeper.GetModuleAccount(ctx, "buyback")
	balances := bankKeeper.GetAllBalances(ctx, buybackacc.GetAddress())
	require.Len(t, balances, 2)
	require.False(t, balances.AmountOf("chf").IsZero())
	require.False(t, balances.AmountOf("eur").IsZero())
	require.True(t, balances.AmountOf("ungm").IsZero())

	// Staking token should be in the distribution account
	distracc := accountKeeper.GetModuleAccount(ctx, authtypes.FeeCollectorName)
	balances = bankKeeper.GetAllBalances(ctx, distracc.GetAddress())

	require.Len(t, balances, 1)
	require.True(t, balances.AmountOf("chf").IsZero())
	require.True(t, balances.AmountOf("eur").IsZero())
	require.False(t, balances.AmountOf("ungm").IsZero())
}

func TestStartTimeInFuture(t *testing.T) {
	ctx, keeper, bankKeeper, _ := createTestComponents()

	initialEurAmount, _ := sdk.ParseCoinNormalized("1000000000eur")
	bankKeeper.SetSupply(ctx, banktypes.NewSupply(sdk.NewCoins(initialEurAmount)))

	ctx = ctx.WithBlockTime(time.Now()).WithBlockHeight(55)

	// Chain is configured (erroneously or not) to start inflation at some point in the future.
	lastAppliedTime := time.Now().Add(2 * time.Hour)
	state := types.InflationState{
		LastAppliedTime:   lastAppliedTime,
		LastAppliedHeight: sdk.ZeroInt(),
		InflationAssets:   nil,
	}

	keeper.SetState(ctx, state)
	keeper.AddDenoms(ctx, []string{"eur"})
	keeper.SetInflation(ctx, sdk.NewDecWithPrec(1, 0), "eur") // 100% inflation

	// Inflation should not have started yet.
	BeginBlocker(ctx, keeper)
	total := bankKeeper.GetSupply(ctx).GetTotal()
	require.Equal(t, initialEurAmount.Amount.String(), total.AmountOf("eur").String())

	// Not yet
	ctx = ctx.WithBlockTime(time.Now().Add(time.Hour)).WithBlockHeight(60)
	BeginBlocker(ctx, keeper)
	total = bankKeeper.GetSupply(ctx).GetTotal()
	require.Equal(t, initialEurAmount.Amount.String(), total.AmountOf("eur").String())

	// Now it should have started increasing the total supply
	ctx = ctx.WithBlockTime(time.Now().Add(3 * time.Hour)).WithBlockHeight(65)
	BeginBlocker(ctx, keeper)
	total = bankKeeper.GetSupply(ctx).GetTotal()
	require.True(t, initialEurAmount.Amount.LT(total.AmountOf("eur")))
}

func createTestComponents() (sdk.Context, keeper.Keeper, bankkeeper.Keeper, authkeeper.AccountKeeper) {
	encConfig := MakeTestEncodingConfig()
	var (
		keyInflation = sdk.NewKVStoreKey(ModuleName)
		bankKey      = sdk.NewKVStoreKey(banktypes.ModuleName)
		authCapKey   = sdk.NewKVStoreKey("authCapKey")
		keyParams    = sdk.NewKVStoreKey("params")
		stakingKey   = sdk.NewKVStoreKey("staking")
		supplyKey    = sdk.NewKVStoreKey("supply")
		authKey      = sdk.NewKVStoreKey(authtypes.StoreKey)
		tkeyParams   = sdk.NewTransientStoreKey("transient_params")

		blacklistedAddrs = make(map[string]bool)
	)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(keyInflation, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(authCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(supplyKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(stakingKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(authKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(bankKey, sdk.StoreTypeIAVL, db)

	err := ms.LoadLatestVersion()
	if err != nil {
		panic(err)
	}

	maccPerms := map[string][]string{
		ModuleName:                     {authtypes.Minter},
		authtypes.FeeCollectorName:     nil,
		"buyback":                      {authtypes.Burner},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
	}

	pk := paramskeeper.NewKeeper(encConfig.Marshaler, encConfig.Amino, keyParams, tkeyParams)

	ctx := sdk.NewContext(ms, tmproto.Header{ChainID: "test-chain"}, true, log.NewNopLogger())

	accountKeeper := authkeeper.NewAccountKeeper(
		encConfig.Marshaler, authCapKey, pk.Subspace(authtypes.ModuleName), authtypes.ProtoBaseAccount, maccPerms,
	)

	bankKeeper := bankkeeper.NewBaseKeeper(
		encConfig.Marshaler, bankKey, accountKeeper, pk.Subspace(banktypes.ModuleName), blacklistedAddrs,
	)

	stakingKeeper := mockStakingKeeper{}

	inflationKeeper := NewKeeper(
		encConfig.Amino, keyInflation, bankKeeper, accountKeeper, stakingKeeper, "buyback", authtypes.FeeCollectorName)

	lastAppliedTime := time.Now().Add(-2400 * time.Hour)

	state := types.InflationState{
		LastAppliedTime:   lastAppliedTime,
		LastAppliedHeight: sdk.ZeroInt(),
		InflationAssets:   nil,
	}
	inflationKeeper.SetState(ctx, state)
	return ctx, inflationKeeper, bankKeeper, accountKeeper
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

func coins(s string) sdk.Coins {
	coins, err := sdk.ParseCoinsNormalized(s)
	if err != nil {
		panic(err)
	}
	return coins
}

type mockStakingKeeper struct{}

func (m mockStakingKeeper) GetParams(_ sdk.Context) stakingtypes.Params {
	return stakingtypes.NewParams(5*time.Minute, 40, 50, 0, "ungm")
}
