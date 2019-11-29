// This software is Copyright (c) 2019 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CodeType = sdk.CodeType

const (
	Codespace sdk.CodespaceType = ModuleName

	//CodeInvalidValidator  CodeType = 101
	//CodeInvalidDelegation CodeType = 102
	//CodeInvalidInput      CodeType = 103
	//CodeValidatorJailed   CodeType = 104
	//CodeInvalidAddress    CodeType = sdk.CodeInvalidAddress
	CodeUnauthorized CodeType = sdk.CodeUnauthorized
	//CodeInternal          CodeType = sdk.CodeInternal
	//CodeUnknownRequest    CodeType = sdk.CodeUnknownRequest
)

func ErrUnauthorizedInflationChange(acc sdk.AccAddress) sdk.Error {
	return sdk.NewError(Codespace, CodeUnauthorized, "Address %v cannot modify inflation", acc)
}
