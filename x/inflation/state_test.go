// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package inflation

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"

	"github.com/e-money/em-ledger/x/inflation/internal/keeper"
	"github.com/e-money/em-ledger/x/inflation/internal/types"
	"github.com/stretchr/testify/require"
	"testing"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

func TestModule1(t *testing.T) {
	ctx, keeper, supplyKeeper := createTestComponents()

	initialEurAmount, _ := sdk.ParseCoin("1000000000eur")
	supplyKeeper.SetSupply(ctx, supply.NewSupply(sdk.NewCoins(initialEurAmount)))

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
		total := supplyKeeper.GetSupply(ctx).GetTotal()
		minted := total.AmountOf("eur").Sub(initialEurAmount.Amount)
		require.True(t, minted.LT(sdk.NewInt(20).MulRaw(i)))
	}
}

func TestStartTimeInFuture(t *testing.T) {
	ctx, keeper, supplyKeeper := createTestComponents()

	initialEurAmount, _ := sdk.ParseCoin("1000000000eur")
	supplyKeeper.SetSupply(ctx, supply.NewSupply(sdk.NewCoins(initialEurAmount)))

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
	total := supplyKeeper.GetSupply(ctx).GetTotal()
	require.Equal(t, initialEurAmount.Amount.String(), total.AmountOf("eur").String())

	// Not yet
	ctx = ctx.WithBlockTime(time.Now().Add(time.Hour)).WithBlockHeight(60)
	BeginBlocker(ctx, keeper)
	total = supplyKeeper.GetSupply(ctx).GetTotal()
	require.Equal(t, initialEurAmount.Amount.String(), total.AmountOf("eur").String())

	// Now it should have started increasing the total supply
	ctx = ctx.WithBlockTime(time.Now().Add(3 * time.Hour)).WithBlockHeight(65)
	BeginBlocker(ctx, keeper)
	total = supplyKeeper.GetSupply(ctx).GetTotal()
	require.True(t, initialEurAmount.Amount.LT(total.AmountOf("eur")))
}

func createTestComponents() (sdk.Context, keeper.Keeper, supply.Keeper) {
	cdc := createCDC()

	var (
		keyInflation = sdk.NewKVStoreKey(ModuleName)
		authCapKey   = sdk.NewKVStoreKey("authCapKey")
		keyParams    = sdk.NewKVStoreKey("params")
		supplyKey    = sdk.NewKVStoreKey("supply")
		tkeyParams   = sdk.NewTransientStoreKey("transient_params")

		blacklistedAddrs = make(map[string]bool)
	)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(keyInflation, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(authCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(supplyKey, sdk.StoreTypeIAVL, db)

	err := ms.LoadLatestVersion()
	if err != nil {
		panic(err)
	}

	pk := params.NewKeeper(cdc, keyParams, tkeyParams, params.DefaultCodespace)

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "test-chain"}, true, log.NewNopLogger())
	accountKeeper := auth.NewAccountKeeper(cdc, authCapKey, pk.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount)

	bankKeeper := bank.NewBaseKeeper(accountKeeper, pk.Subspace(bank.DefaultParamspace), bank.DefaultCodespace, blacklistedAddrs)

	maccperms := map[string][]string{
		ModuleName:            {supply.Minter},
		auth.FeeCollectorName: nil,
	}

	supplyKeeper := supply.NewKeeper(cdc, supplyKey, accountKeeper, bankKeeper, maccperms)

	keeper := NewKeeper(cdc, keyInflation, supplyKeeper, auth.FeeCollectorName)

	lastAppliedTime := time.Now().Add(-2400 * time.Hour)

	state := types.InflationState{
		LastAppliedTime:   lastAppliedTime,
		LastAppliedHeight: sdk.ZeroInt(),
		InflationAssets:   nil,
	}
	keeper.SetState(ctx, state)

	return ctx, keeper, supplyKeeper
}

func createCDC() *codec.Codec {
	cdc := codec.New()
	codec.RegisterCrypto(cdc)
	supply.RegisterCodec(cdc)
	auth.RegisterCodec(cdc)
	bank.RegisterCodec(cdc)
	return cdc

}

func coins(s string) sdk.Coins {
	coins, err := sdk.ParseCoins(s)
	if err != nil {
		panic(err)
	}
	return coins
}
