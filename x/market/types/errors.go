// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrAccountBalanceInsufficient              = sdkerrors.Register(ModuleName, 1, "insufficient account balance")
	ErrAccountBalanceInsufficientForInstrument = sdkerrors.Register(ModuleName, 2, "")
	ErrNonUniqueClientOrderId                  = sdkerrors.Register(ModuleName, 3, "")
	ErrClientOrderIdNotFound                   = sdkerrors.Register(ModuleName, 4, "")
	ErrOrderInstrumentChanged                  = sdkerrors.Register(ModuleName, 5, "")
	ErrInvalidClientOrderId                    = sdkerrors.Register(ModuleName, 6, "")
	ErrInvalidInstrument                       = sdkerrors.Register(ModuleName, 7, "")
	ErrInvalidPrice                            = sdkerrors.Register(ModuleName, 8, "")
	ErrNoSourceRemaining                       = sdkerrors.Register(ModuleName, 9, "")
	ErrUnknownAsset                            = sdkerrors.Register(ModuleName, 10, "")
	ErrUnknownOrderType                        = sdkerrors.Register(ModuleName, 11, "Unknown order type")
	ErrUnknownTimeInForce                      = sdkerrors.Register(ModuleName, 12, "")
	ErrNoMarketDataAvailable                   = sdkerrors.Register(ModuleName, 13, "No market data available for instrument")
	ErrInvalidSlippage                         = sdkerrors.Register(ModuleName, 14, "Invalid slippage")
)
