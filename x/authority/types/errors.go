package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CodeType = sdk.CodeType

const (
	Codespace sdk.CodespaceType = ModuleName

	CodeMissingDenomination sdk.CodeType = 1
	CodeInvalidDenomination sdk.CodeType = 2
)

func ErrNoDenomsSpecified() sdk.Error {
	return sdk.NewError(Codespace, CodeMissingDenomination, "No denominations specified in authority call")
}

func ErrInvalidDenom(denom string) sdk.Error {
	return sdk.NewError(Codespace, CodeInvalidDenomination, "Invalid denomination found: %v", denom)
}
