// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package util

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func ParseDenominations(denoms string) ([]string, error) {
	res := make([]string, 0)
	for _, denom := range strings.Split(denoms, ",") {
		denom = strings.TrimSpace(denom)

		if len(denom) == 0 {
			continue
		}

		if err := sdk.ValidateDenom(denom); err != nil {
			return nil, err
		}

		res = append(res, denom)
	}

	return res, nil
}
