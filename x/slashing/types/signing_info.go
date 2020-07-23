// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Signing info for a validator
type ValidatorSigningInfo struct {
	Address     sdk.ConsAddress `json:"address" yaml:"address"`           // validator consensus address
	JailedUntil time.Time       `json:"jailed_until" yaml:"jailed_until"` // timestamp validator cannot be unjailed until
	Tombstoned  bool            `json:"tombstoned" yaml:"tombstoned"`     // whether or not a validator has been tombstoned (killed out of validator set)
}

// Construct a new `ValidatorSigningInfo` struct
func NewValidatorSigningInfo(condAddr sdk.ConsAddress, jailedUntil time.Time, tombstoned bool) ValidatorSigningInfo {
	return ValidatorSigningInfo{
		Address:     condAddr,
		JailedUntil: jailedUntil,
		Tombstoned:  tombstoned,
	}
}

// Return human readable signing info
func (i ValidatorSigningInfo) String() string {
	return fmt.Sprintf(`Validator Signing Info:
	Address:               %s
	Jailed Until:          %v
	Tombstoned:            %t`,
		i.Address, i.JailedUntil, i.Tombstoned)
}
