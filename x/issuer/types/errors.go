package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

var (
	ErrNegativeMintableBalance     = sdkerrors.Register(ModuleName, 1, "Mintable balance became negative")
	ErrNotLiquidityProvider        = sdkerrors.Register(ModuleName, 2, "Account is not a Liquidity Provider")
	ErrDoesNotControlDenomination  = sdkerrors.Register(ModuleName, 3, "Account is not an Issuer of this Denomination")
	ErrDenominationAlreadyAssigned = sdkerrors.Register(ModuleName, 4, "Domination has already been assigned")
	ErrNotAnIssuer                 = sdkerrors.Register(ModuleName, 5, "Account is not an issuer")
	ErrNegativeInflation           = sdkerrors.Register(ModuleName, 6, "Inflation can't be negative")
	ErrDenomInflation              = sdkerrors.Register(ModuleName, 7, "Inflation denomination error")
)
