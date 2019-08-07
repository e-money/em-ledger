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
	AccrualsPerMinute = 4
	AccrualSlots      = 365 * 24 * 60 * AccrualsPerMinute
)

// BeginBlocker mints new tokens for the previous block.
func BeginBlocker(ctx sdk.Context, k Keeper) {
	// fetch stored minter
	minter := k.GetMinter(ctx)

	// TODO
	// Is it time to accrue interest?
	//t := ctx.BlockTime().Truncate(time.Minute)
	t := truncateTimestamp(ctx.BlockTime(), 60/AccrualsPerMinute)
	if t.Equal(minter.LastAccrual) {
		// A full interest accrual period has not elapsed since last block
		fmt.Println(" *** No accrual for this block.")
		return
	}

	// TODO Calculate number of accrual periods since the last one executed
	// TODO Set a genesis value for minter.LastAccrual

	minter.LastAccrual = t
	k.SetMinter(ctx, minter)

	params := k.GetParams(ctx)

	for _, asset := range params.InflationAssets {
		annualInterest := asset.Inflation

		periodInterest := annualInterest.QuoInt(sdk.NewInt(AccrualSlots))

		supply := k.TotalTokenSupply(ctx, asset.Denom)
		increase := periodInterest.MulInt(supply)

		mintedCoin := sdk.NewCoin(asset.Denom, increase.RoundInt())
		mintedCoins := sdk.NewCoins(mintedCoin)

		err := k.MintCoins(ctx, mintedCoins)
		if err != nil {
			panic(err)
		}

		err = k.AddCollectedFees(ctx, mintedCoins)
		if err != nil {
			panic(err)
		}
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

func truncateTimestamp(ts time.Time, second int) time.Time {
	diff := ts.Second() % second

	ts = ts.Add(-time.Duration(diff) * time.Second)
	ts = ts.Add(-time.Duration(ts.Nanosecond()))

	return ts
}
