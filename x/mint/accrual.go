package mint

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// Apply interest every minute
	//AccrualSlots = 365 * 24 * 60

	// DEBUG value where interest is accrued 4 times per minute
	accrualPeriod = 15 * time.Second
	//accrualPeriod = time.Minute
	AccrualSlots = int64(time.Duration(365*24*time.Hour) / accrualPeriod)
)

// BeginBlocker mints new tokens for the previous block.
func BeginBlocker(ctx sdk.Context, k Keeper) {
	state := k.GetState(ctx)

	// Gate-keep this functionality based on time since last block

	mintedCoins := sdk.Coins{}

	fmt.Println(state)

	totalTokenSupply := k.TotalTokenSupply(ctx)
	for i, asset := range state.InflationAssets {
		supply := totalTokenSupply.AmountOf(asset.Denom)

		currentTime := ctx.BlockTime()
		accum := asset.Accum
		lastAccrual := asset.LastAccrual

		accum, minted := calculateAccrual(accum, supply, asset.Inflation, lastAccrual, currentTime)

		mintedCoins = append(mintedCoins, sdk.NewCoin(asset.Denom, minted))
		asset.Accum = accum
		asset.LastAccrual = ctx.BlockTime()
		state.InflationAssets[i] = asset
	}

	if mintedCoins.IsZero() {
		return
	}

	fmt.Println("MintedCoins:\n", mintedCoins)

	err := k.MintCoins(ctx, mintedCoins)
	if err != nil {
		panic(err)
	}

	err = k.AddMintedCoins(ctx, mintedCoins)
	if err != nil {
		panic(err)
	}

	k.SetState(ctx, state)

	/*
		blocktimeTruncated := truncateTimestamp(ctx.BlockTime(), int(accrualPeriod.Seconds()))
		if blocktimeTruncated.Sub(minter.LastAccrual) < accrualPeriod {
			// A full interest accrual period has not elapsed since last block
			fmt.Println(" *** No accrual for this block.")
			return
		}

		// Determine the number of accrual periods since the last one.
		diff := blocktimeTruncated.Sub(minter.LastAccrual)
		accrualPeriodCount := sdk.NewInt(int64(diff / accrualPeriod))
		fmt.Println(" *** Estimated number of missing accrual periods: ", accrualPeriodCount)

		minter.LastAccrual = blocktimeTruncated
		k.SetMinter(ctx, minter)

		params := k.GetParams(ctx)

		mintedCoins := sdk.Coins{}
		for _, asset := range params.InflationAssets {
			annualInterest := asset.Inflation

			periodInterest := annualInterest.QuoInt(sdk.NewInt(AccrualSlots))

			supply := k.TotalTokenSupply(ctx, asset.Denom)
			increase := periodInterest.MulInt(supply).MulInt(accrualPeriodCount)
			fmt.Printf(" *** Inflating supply of %v by %v%%. Current supply: %v Increase: %v\n'", asset.Denom, periodInterest, supply, increase.RoundInt())

			mintedCoin := sdk.NewCoin(asset.Denom, increase.RoundInt())
			mintedCoins = append(mintedCoins, mintedCoin)
		}

		fmt.Println(" *** Mintedcoins:", mintedCoins)

		err := k.MintCoins(ctx, mintedCoins)
		if err != nil {
			panic(err)
		}

		err = k.AddMintedCoins(ctx, mintedCoins)
		if err != nil {
			panic(err)
		}
	*/
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
