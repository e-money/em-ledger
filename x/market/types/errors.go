// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrAccountBalanceInsufficient              = sdkerrors.Register(ModuleName, 1, "insufficient account balance")
	ErrAccountBalanceInsufficientForInstrument = sdkerrors.Register(ModuleName, 2, "an order exists with the same quantity of source, destination instruments")
	ErrNonUniqueClientOrderId                  = sdkerrors.Register(ModuleName, 3, "the client order id is duplicate and has been used before")
	ErrClientOrderIdNotFound                   = sdkerrors.Register(ModuleName, 4, "the client order cannot be found")
	ErrOrderInstrumentChanged                  = sdkerrors.Register(ModuleName, 5, "cannot change the instrument from the original order")
	ErrInvalidClientOrderId                    = sdkerrors.Register(ModuleName, 6, "the order id length is greater than 32")
	ErrInvalidInstrument                       = sdkerrors.Register(ModuleName, 7, "source and destination instruments are the same")
	ErrInvalidPrice                            = sdkerrors.Register(ModuleName, 8, "insufficient source instrument quantity to pay for 1 unit of destination instrument")
	ErrNoSourceRemaining                       = sdkerrors.Register(ModuleName, 9, "the original order has spent the entire source instrument quantity")
	ErrUnknownAsset                            = sdkerrors.Register(ModuleName, 10, "unknown destination instrument denomination")
	ErrUnknownTimeInForce                      = sdkerrors.Register(ModuleName, 12, "unknown time in force value. Valid values are TimeInForce_GoodTillCancel, TimeInForce_FillOrKill, TimeInForce_ImmediateOrCancel")
	ErrNoMarketDataAvailable                   = sdkerrors.Register(ModuleName, 13, "no market data available for instrument")
	ErrInvalidSlippage                         = sdkerrors.Register(ModuleName, 14, "invalid slippage")
)
