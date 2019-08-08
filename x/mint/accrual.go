package mint

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// Do not accrue interest if less than this period has elapsed since last accrual
	minimumAccrualPeriod = 30 * time.Second
)

// BeginBlocker mints new tokens for the previous block.
func BeginBlocker(ctx sdk.Context, k Keeper) {
	state := k.GetState(ctx)
	blockTime := ctx.BlockTime()

	if blockTime.Sub(state.LastAccrual) < minimumAccrualPeriod {
		return
	}

	totalTokenSupply := k.TotalTokenSupply(ctx)

	// Gate-keep this functionality based on time since last block
	mintedCoins := accrueInterest(&state, totalTokenSupply, blockTime)

	if mintedCoins.IsZero() {
		return
	}

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

func accrueInterest(state *InflationState, totalTokenSupply sdk.Coins, currentTime time.Time) sdk.Coins {
	lastAccrual := state.LastAccrual
	state.LastAccrual = currentTime
	mintedCoins := sdk.Coins{}

	for i, asset := range state.InflationAssets {
		supply := totalTokenSupply.AmountOf(asset.Denom)

		accum, minted := calculateAccrual(asset.Accum, supply, asset.Inflation, lastAccrual, currentTime)

		mintedCoins = append(mintedCoins, sdk.NewCoin(asset.Denom, minted))

		asset.Accum = accum
		state.InflationAssets[i] = asset
	}

	return mintedCoins
}

func calculateAccrual(prevAccum sdk.Dec, supply sdk.Int, annualInterest sdk.Dec, lastAccrual, currentTime time.Time) (accum sdk.Dec, minted sdk.Int) {
	annualNS := time.Duration(365 * 24 * time.Hour).Nanoseconds()

	periodNS := sdk.NewDec(currentTime.Sub(lastAccrual).Nanoseconds())
	accum = annualInterest.MulInt(supply).Mul(periodNS).Add(prevAccum)

	minted = accum.Quo(sdk.NewDec(annualNS)).TruncateInt()
	accum = accum.Sub(minted.MulRaw(annualNS).ToDec())

	return
}

func truncateTimestamp(ts time.Time, second int) time.Time {
	diff := ts.Second() % second

	ts = ts.Add(-time.Duration(diff) * time.Second)
	ts = ts.Add(-time.Duration(ts.Nanosecond()))

	return ts
}
