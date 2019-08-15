package inflation

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// Do not apply inflation if less than this period has elapsed since last accrual
	minimumMintingPeriod = 30 * time.Second
)

// BeginBlocker mints new tokens for the previous block.
func BeginBlocker(ctx sdk.Context, k Keeper) {
	state := k.GetState(ctx)
	blockTime := ctx.BlockTime()

	if blockTime.Sub(state.LastApplied) < minimumMintingPeriod {
		return
	}

	totalTokenSupply := k.TotalTokenSupply(ctx)

	// Gate-keep this functionality based on time since last block
	mintedCoins := applyInflation(&state, totalTokenSupply, blockTime)

	if mintedCoins.IsZero() {
		return
	}

	fmt.Println(" *** Inflation minted coins:", mintedCoins)

	k.SetState(ctx, state)
	err := k.MintCoins(ctx, mintedCoins)
	if err != nil {
		panic(err)
	}

	err = k.AddMintedCoins(ctx, mintedCoins)
	if err != nil {
		panic(err)
	}

	/*
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeMint,
				sdk.NewAttribute(types.AttributeKeyBondedRatio, bondedRatio.String()),
				sdk.NewAttribute(types.AttributeKeyInflation, minter.Inflation.String()),
				sdk.NewAttribute(types.AttributeKeyAnnualProvisions, minter.AnnualProvisions.String()),
				sdk.NewAttribute(types.AttributeKeyAmount, mintedCoin.Amount.String()),
			),
		)
	*/
}

func applyInflation(state *InflationState, totalTokenSupply sdk.Coins, currentTime time.Time) sdk.Coins {
	lastAccrual := state.LastApplied
	state.LastApplied = currentTime
	mintedCoins := sdk.Coins{}

	for i, asset := range state.InflationAssets {
		supply := totalTokenSupply.AmountOf(asset.Denom)

		accum, minted := calculateInflation(asset.Accum, supply, asset.Inflation, lastAccrual, currentTime)

		mintedCoins = append(mintedCoins, sdk.NewCoin(asset.Denom, minted))

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
