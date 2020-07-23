// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const ClientOrderIDMaxLength = 32

var (
	_ sdk.Msg = MsgAddOrder{}
	_ sdk.Msg = MsgCancelOrder{}
	_ sdk.Msg = MsgCancelReplaceOrder{}
)

type (
	MsgAddOrder struct {
		Owner               sdk.AccAddress
		Source, Destination sdk.Coin
		ClientOrderId       string
	}

	MsgCancelOrder struct {
		Owner         sdk.AccAddress
		ClientOrderId string
	}

	MsgCancelReplaceOrder struct {
		Owner                               sdk.AccAddress
		Source, Destination                 sdk.Coin
		OrigClientOrderId, NewClientOrderId string
	}
)

func (m MsgCancelReplaceOrder) Route() string {
	return RouterKey
}

func (m MsgCancelReplaceOrder) Type() string {
	return "cancelreplaceorder"
}

func (m MsgCancelReplaceOrder) ValidateBasic() error {
	if m.Owner.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing owner address")
		//return sdk.ErrInvalidAddress("missing owner address")
	}

	if !m.Destination.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "destination amount is invalid: %v", m.Destination.String())
		//return sdk.ErrInvalidCoins("destination amount is invalid: " + m.Destination.String())
	}

	if !m.Source.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "source amount is invalid: %v", m.Source.String())
		//return sdk.ErrInvalidCoins("source amount is invalid: " + m.Source.String())
	}

	if m.Source.Denom == m.Destination.Denom {
		return sdkerrors.Wrapf(ErrInvalidInstrument, "'%v/%v' is not a valid instrument", m.Source.Denom, m.Destination.Denom)
		//return ErrInvalidInstrument(m.Source.Denom, m.Destination.Denom)
	}

	err := validateClientOrderID(m.OrigClientOrderId)
	if err != nil {
		return err
	}

	return validateClientOrderID(m.NewClientOrderId)
}

func (m MsgCancelReplaceOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

func (m MsgCancelReplaceOrder) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Owner}
}

func (m MsgCancelOrder) Route() string {
	return RouterKey
}

func (m MsgCancelOrder) Type() string {
	return "cancelorder"
}

func (m MsgCancelOrder) ValidateBasic() error {
	if m.Owner.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing owner address")
	}

	return validateClientOrderID(m.ClientOrderId)
}

func (m MsgCancelOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

func (m MsgCancelOrder) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Owner}
}

func (m MsgAddOrder) Route() string {
	return RouterKey
}

func (m MsgAddOrder) Type() string {
	return "addorder"
}

func (m MsgAddOrder) ValidateBasic() error {
	if m.Owner.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing owner address")
		//return sdk.ErrInvalidAddress("missing owner address")
	}

	if !m.Destination.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "destination amount is invalid: %v", m.Destination.String())
		//return sdk.ErrInvalidCoins("destination amount is invalid: " + m.Destination.String())
	}

	if !m.Source.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "source amount is invalid: %v", m.Source.String())
		//return sdk.ErrInvalidCoins("source amount is invalid: " + m.Source.String())
	}

	if m.Source.Denom == m.Destination.Denom {
		return sdkerrors.Wrapf(ErrInvalidInstrument, "'%v/%v' is not a valid instrument", m.Source.Denom, m.Destination.Denom)
		//return ErrInvalidInstrument(m.Source.Denom, m.Destination.Denom)
	}

	return validateClientOrderID(m.ClientOrderId)
}

func (m MsgAddOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

func (m MsgAddOrder) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Owner}
}

func validateClientOrderID(id string) error {
	if len(id) > ClientOrderIDMaxLength {
		return sdkerrors.Wrap(ErrInvalidClientOrderId, id)
		//return ErrInvalidClientOrderId(id)
	}

	return nil
}
