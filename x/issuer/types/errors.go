// This software is Copyright (c) 2019 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

var (
	ErrNegativeMintableBalance     = sdkerrors.Register(ModuleName, 1, "")
	ErrNotLiquidityProvider        = sdkerrors.Register(ModuleName, 2, "")
	ErrDoesNotControlDenomination  = sdkerrors.Register(ModuleName, 3, "")
	ErrDenominationAlreadyAssigned = sdkerrors.Register(ModuleName, 4, "")
	ErrIssuerNotFound              = sdkerrors.Register(ModuleName, 5, "")
	ErrNegativeInflation           = sdkerrors.Register(ModuleName, 6, "")
	ErrNotAnIssuer                 = sdkerrors.Register(ModuleName, 7, "")
)
