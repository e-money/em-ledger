// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package util

import (
	"fmt"
	"strings"

	"github.com/e-money/em-ledger/x/authority/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func ParseDenominations(denoms []string, defDescValue string) ([]types.Denomination, error) {
	if len(denoms) == 0 {
		return []types.Denomination{}, fmt.Errorf("missing denominations")
	}
	res := make([]types.Denomination, 0)
	for _, denom := range denoms {
		denom = strings.Trim(denom, `"'`)
		denomFields := strings.FieldsFunc(strings.TrimSpace(denom), func(r rune) bool {
			return r == ','
		})

		if len(denomFields) == 0 {
			return nil, fmt.Errorf("missing denomination fields")
		}

		if err := sdk.ValidateDenom(denomFields[0]); err != nil {
			return nil, err
		}

		denomStruct := types.Denomination{
			Base:        denomFields[0],
			Display:     strings.ToUpper(denomFields[0]),
			Description: defDescValue,
		}

		if len(denomFields) > 1 {
			denomStruct.Display = denomFields[1]
		}

		if len(denomFields) > 2 {
			denomStruct.Description = denomFields[2]
		}

		res = append(res, denomStruct)
	}

	return res, nil
}
