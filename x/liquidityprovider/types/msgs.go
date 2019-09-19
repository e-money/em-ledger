package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ sdk.Msg = MsgMintTokens{}
)

type MsgMintTokens struct {
	Amount            sdk.Coins
	LiquidityProvider sdk.AccAddress
}

func (msg MsgMintTokens) Route() string { return RouterKey }

func (msg MsgMintTokens) Type() string { return "mint_tokens" }

func (msg MsgMintTokens) ValidateBasic() sdk.Error {
	if msg.LiquidityProvider.Empty() {
		return sdk.ErrInvalidAddress(msg.LiquidityProvider.String())
	}

	if !msg.Amount.IsValid() {
		return sdk.ErrInvalidCoins(msg.Amount.String())
	}

	if msg.Amount.IsAnyNegative() {
		return sdk.ErrInvalidCoins(msg.Amount.String())
	}

	return nil
}

func (msg MsgMintTokens) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg MsgMintTokens) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.LiquidityProvider}
}
