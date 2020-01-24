// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package bank

import sdk "github.com/cosmos/cosmos-sdk/types"

type CodeType = sdk.CodeType

const (
	Codespace sdk.CodespaceType = "embank"

	CodeRestrictedDenomination sdk.CodeType = 1
)

func ErrRestrictedDenominationUsed(denom string) sdk.Error {
	return sdk.NewError(Codespace, CodeRestrictedDenomination, "'%v' is a restricted denomination", denom)
}
