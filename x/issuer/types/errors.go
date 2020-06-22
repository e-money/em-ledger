// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

var (
	ErrNegativeMintableBalance     = sdkerrors.Register(ModuleName, 1, "Mintable balance became negative")
	ErrNotLiquidityProvider        = sdkerrors.Register(ModuleName, 2, "Account is not a Liquidity Provider")
	ErrDoesNotControlDenomination  = sdkerrors.Register(ModuleName, 3, "Not Issuer of this Token Denomination")
	ErrDenominationAlreadyAssigned = sdkerrors.Register(ModuleName, 4, "Token Denomination has already been assigned")
	ErrIssuerNotFound              = sdkerrors.Register(ModuleName, 5, "Issuer does not exist")
	ErrNegativeInflation           = sdkerrors.Register(ModuleName, 6, "")
	ErrNotAnIssuer                 = sdkerrors.Register(ModuleName, 7, "")
)
