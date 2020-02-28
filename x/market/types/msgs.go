// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const ClientOrderIDMaxLength = 32

var (
	_ sdk.Msg = MsgAddOrder{}
	_ sdk.Msg = MsgCancelOrder{}
	_ sdk.Msg = MsgCancelReplaceOrder{}
)

type (
	MsgAddOrder struct {
		Owner         sdk.AccAddress `json:"owner" yaml:"owner"`
		Source        sdk.Coin       `json:"source" yaml:"source"`
		Destination   sdk.Coin       `json:"destination" yaml:"destination"`
		ClientOrderId string         `json:"client_order_id" yaml:"client_order_id"`
	}

	MsgCancelOrder struct {
		Owner         sdk.AccAddress `json:"owner" yaml:"owner"`
		ClientOrderId string         `json:"client_order_id" yaml:"client_order_id"`
	}

	MsgCancelReplaceOrder struct {
		Owner             sdk.AccAddress `json:"owner" yaml:"owner"`
		Source            sdk.Coin       `json:"source" yaml:"source"`
		Destination       sdk.Coin       `json:"destination" yaml:"destination"`
		OrigClientOrderId string         `json:"original_client_order_id" yaml:"original_client_order_id"`
		NewClientOrderId  string         `json:"client_order_id" yaml:"client_order_id"`
	}
)

func (m MsgCancelReplaceOrder) Route() string {
	return RouterKey
}

func (m MsgCancelReplaceOrder) Type() string {
	return "cancelreplaceorder"
}

func (m MsgCancelReplaceOrder) ValidateBasic() sdk.Error {
	if m.Owner.Empty() {
		return sdk.ErrInvalidAddress("missing owner address")
	}

	if !m.Destination.IsValid() {
		return sdk.ErrInvalidCoins("destination amount is invalid: " + m.Destination.String())
	}

	if !m.Source.IsValid() {
		return sdk.ErrInvalidCoins("source amount is invalid: " + m.Source.String())
	}

	if m.Source.Denom == m.Destination.Denom {
		return ErrInvalidInstrument(m.Source.Denom, m.Destination.Denom)
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

func (m MsgCancelOrder) ValidateBasic() sdk.Error {
	if m.Owner.Empty() {
		return sdk.ErrInvalidAddress("missing owner address")
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

func (m MsgAddOrder) ValidateBasic() sdk.Error {
	if m.Owner.Empty() {
		return sdk.ErrInvalidAddress("missing owner address")
	}

	if !m.Destination.IsValid() {
		return sdk.ErrInvalidCoins("destination amount is invalid: " + m.Destination.String())
	}

	if !m.Source.IsValid() {
		return sdk.ErrInvalidCoins("source amount is invalid: " + m.Source.String())
	}

	if m.Source.Denom == m.Destination.Denom {
		return ErrInvalidInstrument(m.Source.Denom, m.Destination.Denom)
	}

	return validateClientOrderID(m.ClientOrderId)
}

func (m MsgAddOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

func (m MsgAddOrder) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Owner}
}

func validateClientOrderID(id string) sdk.Error {
	if len(id) > ClientOrderIDMaxLength {
		return ErrInvalidClientOrderId(id)
	}

	return nil
}
