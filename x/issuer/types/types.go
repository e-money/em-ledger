// This software is Copyright (c) 2019 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Issuers []Issuer

type Issuer struct {
	Address sdk.AccAddress `json:"address" yaml:"address"`
	Denoms  []string       `json:"denoms" yaml:"denoms"`
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

func (i Issuers) String() string {
	var sb strings.Builder

	for _, issuer := range i {
		sb.WriteString(fmt.Sprintf("%v : %v\n", issuer.Address.String(), issuer.Denoms))
	}

	return sb.String()
}
