// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = MsgIncreaseMintable{}
	_ sdk.Msg = MsgDecreaseMintable{}
	_ sdk.Msg = MsgRevokeLiquidityProvider{}
	_ sdk.Msg = MsgSetInflation{}
)

type (
	MsgIncreaseMintable struct {
		Issuer            sdk.AccAddress `json:"issuer" yaml:"issuer"`
		LiquidityProvider sdk.AccAddress `json:"liquidity_provider" yaml:"liquidity_provider"`
		MintableIncrease  sdk.Coins      `json:"amount" yaml:"amount"`
	}

	MsgDecreaseMintable struct {
		Issuer            sdk.AccAddress `json:"issuer" yaml:"issuer"`
		LiquidityProvider sdk.AccAddress `json:"liquidity_provider" yaml:"liquidity_provider"`
		MintableDecrease  sdk.Coins      `json:"amount" yaml:"amount"`
	}

	MsgRevokeLiquidityProvider struct {
		Issuer            sdk.AccAddress `json:"issuer" yaml:"issuer"`
		LiquidityProvider sdk.AccAddress `json:"liquidity_provider" yaml:"liquidity_provider"`
	}

	MsgSetInflation struct {
		Issuer        sdk.AccAddress `json:"issuer" yaml:"issuer"`
		Denom         string         `json:"denom" yaml:"denom"`
		InflationRate sdk.Dec        `json:"inflation_rate" yaml:"inflation_rate"`
	}
)

func (msg MsgSetInflation) Route() string { return ModuleName }

func (msg MsgSetInflation) Type() string { return "set_inflation" }

func (msg MsgSetInflation) ValidateBasic() error {
	if msg.InflationRate.IsNegative() {
		return sdkerrors.Wrap(ErrNegativeInflation, "cannot set negative inflation")
		//return ErrNegativeInflation()
	}

	if msg.Issuer.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing issuer address")
		//return sdk.ErrInvalidAddress("missing issuer address")
	}

	return nil
}

func (msg MsgSetInflation) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg MsgSetInflation) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Issuer}
}

func (msg MsgRevokeLiquidityProvider) Route() string { return ModuleName }

func (msg MsgRevokeLiquidityProvider) Type() string { return "revoke_liquidity_provider" }

func (msg MsgRevokeLiquidityProvider) ValidateBasic() error {
	if msg.LiquidityProvider.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing liquidity provider address")
		//return sdk.ErrInvalidAddress("missing liquidity provider address")
	}

	if msg.Issuer.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing issuer address")
		//return sdk.ErrInvalidAddress("missing issuer address")
	}

	return nil
}

func (msg MsgRevokeLiquidityProvider) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg MsgRevokeLiquidityProvider) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Issuer}
}

func (msg MsgDecreaseMintable) Route() string { return ModuleName }

func (msg MsgDecreaseMintable) Type() string { return "decrease_mintable" }

func (msg MsgDecreaseMintable) ValidateBasic() error {
	if msg.LiquidityProvider.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing liquidity provider address")
		//return sdk.ErrInvalidAddress("missing liquidity provider address")
	}

	if msg.Issuer.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing issuer address")
		//return sdk.ErrInvalidAddress("missing issuer address")
	}

	if !msg.MintableDecrease.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "requested decrease is invalid: %v", msg.MintableDecrease.String())
		//return sdk.ErrInvalidCoins("requested decrease is invalid: " + msg.MintableDecrease.String())
	}

	return nil
}

func (msg MsgDecreaseMintable) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg MsgDecreaseMintable) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Issuer}
}

func (msg MsgIncreaseMintable) Route() string { return ModuleName }

func (msg MsgIncreaseMintable) Type() string { return "increase_mintable" }

func (msg MsgIncreaseMintable) ValidateBasic() error {
	if msg.LiquidityProvider.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing liquidity provider address")
		//return sdk.ErrInvalidAddress("missing liquidity provider address")
	}

	if msg.Issuer.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing issuer address")
		//return sdk.ErrInvalidAddress("missing issuer address")
	}

	if !msg.MintableIncrease.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "mintable increase is invalid: "+msg.MintableIncrease.String())
		//return sdk.ErrInvalidCoins("mintable increase is invalid: " + msg.MintableIncrease.String())
	}

	return nil
}

func (msg MsgIncreaseMintable) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg MsgIncreaseMintable) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Issuer}
}
