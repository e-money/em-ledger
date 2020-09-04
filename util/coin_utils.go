package util

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Split a list of coins into two collections: Selected denominations and remaining denominations
func SplitCoinsByDenom(in sdk.Coins, denoms ...string) (selected sdk.Coins, remaining sdk.Coins) {
	remaining = in
	selected = sdk.NewCoins()

	for _, denom := range denoms {
		amount := in.AmountOf(denom)
		if amount.IsZero() {
			continue
		}

		selected = selected.Add(sdk.NewCoin(denom, amount))
	}

	remaining = remaining.Sub(selected)
	return
}
