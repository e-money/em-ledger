// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package util

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func ValidateDenom(denom string) bool {
	err := sdk.ValidateDenom(denom)
	return err == nil
}

func ParseDenominations(denoms string) ([]string, error) {
	res := make([]string, 0)
	for _, denom := range strings.Split(denoms, ",") {
		denom = strings.TrimSpace(denom)

		if len(denom) == 0 {
			continue
		}

		if !ValidateDenom(denom) {
			return nil, fmt.Errorf("invalid denomination: %v", denom)
		}

		res = append(res, denom)
	}

	return res, nil
}
