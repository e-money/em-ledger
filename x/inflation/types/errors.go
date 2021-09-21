// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

var (
	ErrInvalidValidator  = sdkerrors.Register(ModuleName, 1, "")
	ErrInvalidDelegation = sdkerrors.Register(ModuleName, 2, "")
	ErrInvalidInput      = sdkerrors.Register(ModuleName, 3, "")
	ErrValidatorJailed   = sdkerrors.Register(ModuleName, 4, "")
	ErrInvalidAddress    = sdkerrors.Register(ModuleName, 5, "")
	ErrUnauthorized      = sdkerrors.Register(ModuleName, 6, "")
	ErrInternal          = sdkerrors.Register(ModuleName, 7, "")
	ErrUnknownRequest    = sdkerrors.Register(ModuleName, 8, "")
)
