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
