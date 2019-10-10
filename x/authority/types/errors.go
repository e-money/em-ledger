package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CodeType = sdk.CodeType

const (
	Codespace sdk.CodespaceType = ModuleName

	CodeNotAuthority        sdk.CodeType = 1
	CodeMissingDenomination sdk.CodeType = 2
	CodeInvalidDenomination sdk.CodeType = 3
	CodeNoAuthority         sdk.CodeType = 4
)

func ErrNotAuthority(address string) sdk.Error {
	return sdk.NewError(Codespace, CodeMissingDenomination, "No denominations specified in authority call")
}

func ErrNoDenomsSpecified() sdk.Error {
	return sdk.NewError(Codespace, CodeMissingDenomination, "No denominations specified in authority call")
}

func ErrInvalidDenom(denom string) sdk.Error {
	return sdk.NewError(Codespace, CodeInvalidDenomination, "Invalid denomination found: %v", denom)
}
func ErrNoAuthorityConfigured() sdk.Error {
	return sdk.NewError(Codespace, CodeNoAuthority, "No authority configured")
}
