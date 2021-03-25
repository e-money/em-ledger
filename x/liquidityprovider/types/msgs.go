// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = &MsgMintTokens{}
	_ sdk.Msg = &MsgBurnTokens{}
)

func (msg MsgBurnTokens) Route() string { return RouterKey }

func (msg MsgBurnTokens) Type() string { return "burn_tokens" }

func (msg MsgBurnTokens) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.LiquidityProvider); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid liquidity provider address (%s)", err)
	}

	if !msg.Amount.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.Amount.String())
		//return sdk.ErrInvalidCoins(msg.Amount.String())
	}

	return nil
}

func (msg MsgBurnTokens) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgBurnTokens) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.LiquidityProvider)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg MsgMintTokens) Route() string { return RouterKey }

func (msg MsgMintTokens) Type() string { return "mint_tokens" }

func (msg MsgMintTokens) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.LiquidityProvider); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid liquidity provider address (%s)", err)
	}

	if !msg.Amount.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.Amount.String())
		//return sdk.ErrInvalidCoins(msg.Amount.String())
	}

	return nil
}

func (msg MsgMintTokens) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgMintTokens) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.LiquidityProvider)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}
