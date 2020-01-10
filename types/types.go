package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type RestrictedDenom struct {
	Denom   string
	Allowed []sdk.AccAddress
}

type RestrictedDenoms []RestrictedDenom

func (rd RestrictedDenoms) Find(denom string) (RestrictedDenom, bool) {
	for _, d := range rd {
		if d.Denom == denom {
			return d, true
		}
	}

	return RestrictedDenom{}, false
}

func (r RestrictedDenom) IsAnyAllowed(in ...sdk.AccAddress) bool {
	for _, addr := range r.Allowed {
		for _, inaddr := range in {
			if inaddr.Equals(addr) {
				return true
			}
		}
	}

	return false
}
