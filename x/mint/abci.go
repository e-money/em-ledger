package mint

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	accrualSlots = 365 * 24 * 60
)

// BeginBlocker mints new tokens for the previous block.
func BeginBlocker(ctx sdk.Context, k Keeper) {
	// Temporary calculation of interest accrual for a 1-minute slot:

	annualInterest := sdk.NewDec(1)
	periodInterest := annualInterest.QuoInt(sdk.NewInt(accrualSlots))
	fmt.Println("PeriodInterest determined to be : ", periodInterest)

	// fetch stored minter & params
	minter := k.GetMinter(ctx)
	params := k.GetParams(ctx)

	// recalculate inflation rate
	//totalStakingSupply := k.StakingTokenSupply(ctx)
	//fmt.Println(" *** TotalStakingSupply", totalStakingSupply)
	//bondedRatio := k.BondedRatio(ctx)

	var modifiedMinter bool
	//nextInflation := minter.NextInflationRate(params, bondedRatio)
	//
	//if !nextInflation.Equal(minter.Inflation) {
	//	fmt.Println("Inflation: ", nextInflation, minter.Inflation)
	//	minter.Inflation = nextInflation
	//	modifiedMinter = true
	//}
	//
	//nextAnnualProvisions := minter.NextAnnualProvisions(params, totalStakingSupply)
	//
	//if !nextAnnualProvisions.Equal(minter.AnnualProvisions) {
	//	fmt.Println("Annual provisions: ", nextAnnualProvisions, minter.AnnualProvisions)
	//	minter.AnnualProvisions = nextAnnualProvisions
	//	modifiedMinter = true
	//}

	// TODO
	// Is it time to accrue interest?
	t := ctx.BlockTime().Truncate(time.Minute)
	fmt.Println(" *** Last minute", t)

	if !t.Equal(minter.LastAccrual) {
		minter.LastAccrual = t
		modifiedMinter = true
	}

	if modifiedMinter {
		fmt.Println(" *** Minter modified!", minter.LastAccrual)
		k.SetMinter(ctx, minter)
	}

	// mint coins, update supply
	mintedCoin := minter.BlockProvision(params)

	if mintedCoin.Amount.IsZero() {
		//fmt.Println(" *** No new coins minted")
		return
	}

	mintedCoins := sdk.NewCoins(mintedCoin)
	fmt.Println(" *** mintedCoins", mintedCoins)

	err := k.MintCoins(ctx, mintedCoins)
	if err != nil {
		panic(err)
	}

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
