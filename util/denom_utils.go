// This software is Copyright (c) 2019 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package util

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"strings"
)

// The denomination validation functions are buried deep inside the Coin struct, so use this approach to validate names.
func ValidateDenom(denomination string) bool {
	defer func() {
		recover()
	}()
	// Function panics when encountering an invalid denomination
	sdk.NewCoin(denomination, sdk.ZeroInt())
	return true
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
