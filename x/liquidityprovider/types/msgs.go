package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ sdk.Msg = MsgMintTokens{}
	_ sdk.Msg = MsgDevTracerBullet{}
)

// DEBUG message that can be used to trigger temporary functionality
type MsgDevTracerBullet struct {
	Sender sdk.AccAddress
}

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

func (msg MsgDevTracerBullet) Route() string { return RouterKey }

func (msg MsgDevTracerBullet) Type() string {
	return "DEBUG"
}

func (msg MsgDevTracerBullet) ValidateBasic() sdk.Error {
	return nil
}

func (msg MsgDevTracerBullet) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg MsgDevTracerBullet) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
