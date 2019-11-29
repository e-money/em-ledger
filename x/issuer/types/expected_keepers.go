// This software is Copyright (c) 2019 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type InflationKeeper interface {
	SetInflation(sdk.Context, sdk.Dec, string) sdk.Result
	AddDenoms(sdk.Context, []string) sdk.Result
}
