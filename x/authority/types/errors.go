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

	CodeNotAuthority        sdk.CodeType = 1
	CodeMissingDenomination sdk.CodeType = 2
	CodeInvalidDenomination sdk.CodeType = 3
	CodeNoAuthority         sdk.CodeType = 4
	CodeInvalidGasPrices    sdk.CodeType = 5
	CodeUnknownDenomination sdk.CodeType = 6
)

func ErrNotAuthority(address string) sdk.Error {
	return sdk.NewError(Codespace, CodeMissingDenomination, "%v is not the authority", address)
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

func ErrInvalidGasPrices(amt string) sdk.Error {
	return sdk.NewError(Codespace, CodeInvalidGasPrices, "Invalid gas prices : %v", amt)
}

func ErrUnknownDenom(denom string) sdk.Error {
	return sdk.NewError(Codespace, CodeUnknownDenomination, "Unknown denomination specified: %v", denom)
}
