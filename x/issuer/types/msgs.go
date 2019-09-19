package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ sdk.Msg = MsgIncreaseCredit{}
)

// Increase the credit of a liquidity provider. If the account is not previously an LP, it will be made one.
type MsgIncreaseCredit struct {
	CreditIncrease    sdk.Coins
	LiquidityProvider sdk.AccAddress
	Issuer            sdk.AccAddress
}

func (msg MsgIncreaseCredit) Route() string { return ModuleName }

func (msg MsgIncreaseCredit) Type() string { return "increaseCredit" }

func (msg MsgIncreaseCredit) ValidateBasic() sdk.Error {
	if msg.LiquidityProvider.Empty() {
		return sdk.ErrInvalidAddress("missing liquidity provider address")
	}

	if msg.Issuer.Empty() {
		return sdk.ErrInvalidAddress("missing issuer address")
	}

	if !msg.CreditIncrease.IsValid() {
		return sdk.ErrInvalidCoins("credit increase is invalid: " + msg.CreditIncrease.String())
	}

	fmt.Println(" *** Message is valid!")

	return nil
}

func (msg MsgIncreaseCredit) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg MsgIncreaseCredit) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Issuer}
}
