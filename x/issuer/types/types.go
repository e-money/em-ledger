package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"sort"
)

type Issuer struct {
	Address sdk.AccAddress
	Denoms  []string
}

func NewIssuer(address sdk.AccAddress, denoms ...string) Issuer {
	sort.Strings(denoms)

	return Issuer{
		Address: address,
		Denoms:  denoms,
	}
}

func (i Issuer) IsValid() bool {
	if i.Address == nil {
		return false
	}

	if len(i.Denoms) == 0 {
		return false
	}

	return true
}
