package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultCodespace sdk.CodespaceType = "iss"

	CodeNegativeMintable      sdk.CodeType = 1
	CodeNotLiquidityProvider  sdk.CodeType = 2
	CodeDuplicateDenomination sdk.CodeType = 3
	CodeIssuerNotFound        sdk.CodeType = 4
	CodeNegativeInflation     sdk.CodeType = 5
	CodeDoesNotControlDenom   sdk.CodeType = 6
	CodeNotAnIssuer           sdk.CodeType = 7
)

func ErrNotAnIssuer(address sdk.AccAddress) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeNotAnIssuer, fmt.Sprintf("%v is not an issuer", address))
}

func ErrNegativeMintableBalance(lp sdk.AccAddress) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeNegativeMintable, fmt.Sprintf("mintable balance decrease would become negative for %d", lp))
}

func ErrNotLiquidityProvider(lp sdk.AccAddress) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeNotLiquidityProvider, fmt.Sprint("account is not a liquidity provider:", lp))
}

func ErrDoesNotControlDenomination(denom string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeDoesNotControlDenom, fmt.Sprintf("issuer does not control inflation of denomination %v", denom))
}

func ErrDenominationAlreadyAssigned() sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeDuplicateDenomination, "denomination is already under control of an issuer")
}

func ErrIssuerNotFound(issuer sdk.AccAddress) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeIssuerNotFound, fmt.Sprintf("unable to find issuer %v", issuer))
}

func ErrNegativeInflation() sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeNegativeInflation, "cannot set negative inflation")
}
