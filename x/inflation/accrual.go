// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package inflation

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/util"
	"github.com/e-money/em-ledger/x/inflation/types"
)

const (
	// Do not apply inflation if less than this period has elapsed since last accrual
	minimumMintingPeriod = 10 * time.Second
)

// BeginBlocker mints new tokens for the previous block.
func BeginBlocker(ctx sdk.Context, k Keeper) {
	state := k.GetState(ctx)
	blockTime := ctx.BlockTime()

	// Gate-keep this functionality based on time since last block to prevent a cascade of blocks
	if blockTime.Sub(state.LastAppliedTime) < minimumMintingPeriod {
		return
	}

	if ctx.BlockHeight() == state.LastAppliedHeight.Int64()+1 {
		return
	}

	// Inflation may be set to start in the future. Do nothing in that case.
	if blockTime.Before(state.LastAppliedTime) {
		return
	}

	totalTokenSupply, err := k.TotalTokenSupply(ctx)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("Inflation module error from bank.GetPaginatedTotalSupply() %v", err))
		return
	}

	mintedCoins := applyInflation(&state, totalTokenSupply, blockTime)
	state.LastAppliedHeight = sdk.NewInt(ctx.BlockHeight())

	k.SetState(ctx, state)

	if mintedCoins.IsZero() {
		return
	}

	k.Logger(ctx).Info("Inflation minted coins", toKeyValuePairs(mintedCoins)...)

	err = k.MintCoins(ctx, mintedCoins)
	if err != nil {
		panic(err)
	}

	// Divide into two pools: Staking tokens and Stablecoin tokens
	stakingDenom := k.GetStakingDenomination(ctx)
	stakingTokens, coinTokens := util.SplitCoinsByDenom(mintedCoins, stakingDenom)

	err = k.DistributeMintedCoins(ctx, coinTokens)
	if err != nil {
		panic(err)
	}

	err = k.DistributeStakingCoins(ctx, stakingTokens)
	if err != nil {
		panic(err)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeInflation,
			sdk.NewAttribute(types.AttributeKeyAction, "mint"),
			sdk.NewAttribute(types.AttributeKeyAmount, mintedCoins.String()),
		),
	)
}

func applyInflation(state *InflationState, totalTokenSupply sdk.Coins, currentTime time.Time) sdk.Coins {
	lastAccrual := state.LastAppliedTime
	mintedCoins := sdk.Coins{}

	state.LastAppliedTime = currentTime

	for i, asset := range state.InflationAssets {
		supply := totalTokenSupply.AmountOf(asset.Denom)

		accum, minted := calculateInflation(asset.Accum, supply, asset.Inflation, lastAccrual, currentTime)

		if minted.IsPositive() { // Coins.IsValid() considers any coin of amount 0 to be invalid, so filter 0 coins.
			mintedCoins = append(mintedCoins, sdk.NewCoin(asset.Denom, minted))
		}

		asset.Accum = accum
		state.InflationAssets[i] = asset
	}

	return mintedCoins.Sort()
}

func calculateInflation(prevAccum sdk.Dec, supply sdk.Int, annualInflation sdk.Dec, lastAccrual, currentTime time.Time) (accum sdk.Dec, minted sdk.Int) {
	annualNS := 365 * 24 * time.Hour.Nanoseconds()

	periodNS := sdk.NewDec(currentTime.Sub(lastAccrual).Nanoseconds())
	accum = annualInflation.MulInt(supply).Mul(periodNS).Add(prevAccum)

	minted = accum.Quo(sdk.NewDec(annualNS)).TruncateInt()
	accum = accum.Sub(minted.MulRaw(annualNS).ToDec())

	return
}

// For use in logging
func toKeyValuePairs(coins sdk.Coins) (res []interface{}) {
	for _, coin := range coins {
		res = append(res, coin.Denom, coin.Amount.String())
	}
	return
}
