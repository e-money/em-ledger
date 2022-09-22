// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrNotAuthority          = sdkerrors.Register(ModuleName, 1, "not an authority")
	ErrNoDenomsSpecified     = sdkerrors.Register(ModuleName, 2, "No denominations specified in authority call")
	ErrInvalidDenom          = sdkerrors.Register(ModuleName, 3, "Invalid denomination found")
	ErrNoAuthorityConfigured = sdkerrors.Register(ModuleName, 4, "No authority configured")
	ErrInvalidGasPrices      = sdkerrors.Register(ModuleName, 5, "Invalid gas prices")
	ErrUnknownDenom          = sdkerrors.Register(ModuleName, 6, "Unknown denomination specified")
	ErrMissingFlag           = sdkerrors.Register(ModuleName, 7, "missing flag")
	ErrGetTotalSupply        = sdkerrors.Register(ModuleName, 8, "GetPaginatedSupply() erred")
	//	ErrPlanTimeIsSet         = sdkerrors.Register(ModuleName, 8, "upgrade plan cannot set time")
	ErrNoParams = sdkerrors.Register(ModuleName, 9, "no parameter changes specified")
)
