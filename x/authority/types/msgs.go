// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = MsgCreateIssuer{}
	_ sdk.Msg = MsgDestroyIssuer{}
	_ sdk.Msg = MsgSetGasPrices{}
)

type (
	MsgCreateIssuer struct {
		Issuer        sdk.AccAddress
		Denominations []string
		Authority     sdk.AccAddress
	}
	MsgDestroyIssuer struct {
		Issuer    sdk.AccAddress
		Authority sdk.AccAddress
	}

	MsgSetGasPrices struct {
		GasPrices sdk.DecCoins
		Authority sdk.AccAddress
	}
)

func (msg MsgDestroyIssuer) Type() string { return "destroy_issuer" }

func (msg MsgCreateIssuer) Type() string { return "create_issuer" }

func (msg MsgSetGasPrices) Type() string { return "set_gas_prices" }

func (msg MsgDestroyIssuer) ValidateBasic() error {
	if msg.Issuer.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "Missing issuer address")
		//return sdk.ErrInvalidAddress("missing issuer address")
	}

	if msg.Authority.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "Missing authority address")
		//return sdk.ErrInvalidAddress("missing authority address")
	}

	return nil
}

func (msg MsgCreateIssuer) ValidateBasic() error {
	if msg.Issuer.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "Missing issuer address")
		//return sdk.ErrInvalidAddress("missing issuer address")
	}

	if msg.Authority.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "Missing authority address")
		//return sdk.ErrInvalidAddress("missing authority address")
	}

	if len(msg.Denominations) == 0 {
		return sdkerrors.Wrap(ErrNoDenomsSpecified, "No denomination specified")
		//return ErrNoDenomsSpecified()
	}

	return nil
}

func (msg MsgSetGasPrices) ValidateBasic() error {
	if msg.Authority.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "Missing authority address")
		//return sdk.ErrInvalidAddress("missing authority address")
	}

	if !msg.GasPrices.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "%v", msg.GasPrices)
		//return sdk.ErrInvalidCoins(msg.GasPrices.String())
	}

	return nil
}

func (msg MsgDestroyIssuer) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Authority}
}

func (msg MsgCreateIssuer) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Authority}
}

func (msg MsgSetGasPrices) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Authority}
}

func (msg MsgDestroyIssuer) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg MsgCreateIssuer) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg MsgSetGasPrices) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg MsgDestroyIssuer) Route() string { return ModuleName }

func (msg MsgCreateIssuer) Route() string { return ModuleName }

func (msg MsgSetGasPrices) Route() string { return ModuleName }
