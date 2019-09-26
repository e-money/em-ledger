package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type InflationKeeper interface {
	SetInflation(sdk.Context, sdk.Dec, string) sdk.Error
}
