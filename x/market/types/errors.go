// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CodeType = sdk.CodeType

const (
	Codespace sdk.CodespaceType = ModuleName

	CodeInsufficientBalance    sdk.CodeType = 1
	CodeNonUniqueClientOrderId sdk.CodeType = 2
	CodeClientOrderIdNotFound  sdk.CodeType = 3
	CodeOrderInstrumentChanged sdk.CodeType = 4
	CodeInvalidClientOrderId   sdk.CodeType = 5
	CodeInvalidInstrument      sdk.CodeType = 6
	CodeInvalidOrderPrice      sdk.CodeType = 7
	CodeInvalidOrder           sdk.CodeType = 8
	CodeUnknownAsset           sdk.CodeType = 9
)

func ErrAccountBalanceInsufficient(address sdk.AccAddress, required sdk.Coin, balance sdk.Int) sdk.Error {
	return sdk.NewError(Codespace, CodeInsufficientBalance, "Account %v has insufficient balance to execute trade: %v < %v", address.String(), balance, required)
}

func ErrAccountBalanceInsufficientForInstrument(address sdk.AccAddress, required sdk.Coin, balance sdk.Int, dst string) sdk.Error {
	return sdk.NewError(Codespace, CodeInsufficientBalance, "Account %v has insufficient balance to execute all orders on instrument %v/%v: %v < %v", address.String(), required.Denom, dst, balance, required)
}

func ErrNonUniqueClientOrderId(address sdk.AccAddress, clientOrderId string) sdk.Error {
	return sdk.NewError(Codespace, CodeNonUniqueClientOrderId, "Account %v already has an active order with client order id: %v", address.String(), clientOrderId)
}

func ErrClientOrderIdNotFound(address sdk.AccAddress, clientOrderId string) sdk.Error {
	return sdk.NewError(Codespace, CodeClientOrderIdNotFound, "Account %v does not have an active order with client order id: %v", address.String(), clientOrderId)
}

func ErrOrderInstrumentChanged(origSrc, origDst, newSrc, newDst string) sdk.Error {
	return sdk.NewError(Codespace, CodeOrderInstrumentChanged, "Instrument cannot be changed when using CancelReplace : %v -> %v != %v -> %v", origSrc, origDst, newSrc, newDst)
}

func ErrInvalidClientOrderId(clientorderid string) sdk.Error {
	return sdk.NewError(Codespace, CodeInvalidClientOrderId, "Specified client order ID is not valid: '%v'", clientorderid)
}

func ErrInvalidInstrument(src, dst string) sdk.Error {
	return sdk.NewError(Codespace, CodeInvalidInstrument, "'%v/%v' is not a valid instrument", src, dst)
}

func ErrInvalidPrice(src, dst sdk.Coin) sdk.Error {
	return sdk.NewError(Codespace, CodeInvalidOrderPrice, "Order price is invalid: %s -> %s", src, dst)
}

func ErrInvalidOrder(order *Order) sdk.Error {
	return sdk.NewError(Codespace, CodeInvalidOrder, "Order is not valid: %v", order)
}

func ErrUnknownAsset(coin sdk.Coin) sdk.Error {
	return sdk.NewError(Codespace, CodeUnknownAsset, "'%v' is not a known asset.", coin.Denom)
}
