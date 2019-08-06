package mint

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// Apply interest every minute
	AccrualSlots = 365 * 24 * 60
)

// BeginBlocker mints new tokens for the previous block.
func BeginBlocker(ctx sdk.Context, k Keeper) {
	// fetch stored minter & params
	minter := k.GetMinter(ctx)
	//params := k.GetParams(ctx)

	// TODO
	// Is it time to accrue interest?
	t := ctx.BlockTime().Truncate(time.Minute)
	if t.Equal(minter.LastAccrual) {
		// A full interest accrual period has not elapsed since last block
		fmt.Println(" *** No accrual for this block.")
		return
	}

	fmt.Println(" *** Minute time difference for block height: ", minter.LastAccrual.Minute(), ctx.BlockTime().Minute(), ctx.BlockHeight())

	// TODO Calculate number of accrual periods since the last one executed
	duration := t.Sub(minter.LastAccrual)
	fmt.Println(" *** Minutes since last accrual", int(duration.Minutes()))
	// TODO Set a genesis value for minter.LastAccrual

	minter.LastAccrual = t
	k.SetMinter(ctx, minter)

	// mint coins, update supply
	annualInterest := sdk.NewDecWithPrec(1, 2) // TODO Store in params
	periodInterest := annualInterest.QuoInt(sdk.NewInt(AccrualSlots))

	ungmSupply := k.TotalTokenSupply(ctx, "ungm")
	fmt.Println(" *** Total ungm supply", ungmSupply)

	coinIncrease := periodInterest.MulInt(ungmSupply)

	mintedCoin := sdk.NewCoin("ungm", coinIncrease.RoundInt())
	mintedCoins := sdk.NewCoins(mintedCoin)

	fmt.Println(" *** mintedCoins", mintedCoins)

	err := k.MintCoins(ctx, mintedCoins)
	if err != nil {
		panic(err)
	}

	ungmSupply = k.TotalTokenSupply(ctx, "ungm")
	fmt.Println(" *** Total supply after accrual", ungmSupply)

	// send the minted coins to the fee collector account
	err = k.AddCollectedFees(ctx, mintedCoins)
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
