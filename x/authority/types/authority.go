// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (rd RestrictedDenoms) Find(denom string) (RestrictedDenom, bool) {
	for _, d := range rd.Denoms {
		if d.Denom == denom {
			return d, true
		}
	}

	return RestrictedDenom{}, false
}

func (r RestrictedDenom) IsAnyAllowed(in ...sdk.AccAddress) bool {
	for _, addr := range r.Allowed {
		for _, inaddr := range in {
			if addr == inaddr.String() {
				return true
			}
		}
	}

	return false
}
