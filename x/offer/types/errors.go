package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CodeType = sdk.CodeType

const (
	Codespace sdk.CodespaceType = ModuleName

	CodeInsufficientBalance    sdk.CodeType = 1
	CodeNonUniqueClientOrderId sdk.CodeType = 2
)

func ErrAccountBalanceInsufficient(address sdk.AccAddress, required sdk.Coin, balance sdk.Int) sdk.Error {
	return sdk.NewError(Codespace, CodeInsufficientBalance, "Account %v has insufficient balance to execute trade: %v < %v", address.String(), balance, required)
}

func ErrNonUniqueClientOrderID(address sdk.AccAddress, clientOrderId string) sdk.Error {
	return sdk.NewError(Codespace, CodeNonUniqueClientOrderId, "Account %v already has an active order with client order id: %v", address.String(), clientOrderId)
}
