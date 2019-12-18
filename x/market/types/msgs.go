// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _, _, _ sdk.Msg = MsgAddOrder{}, MsgCancelOrder{}, MsgCancelReplaceOrder{}

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
		Owner               sdk.AccAddress
		ClientOrderId       string
		Source, Destination sdk.Coin
	}
)

func (m MsgCancelReplaceOrder) Route() string {
	return RouterKey
}

func (m MsgCancelReplaceOrder) Type() string {
	return "cancelreplaceorder"
}

func (m MsgCancelReplaceOrder) ValidateBasic() sdk.Error {
	panic("implement me")
}

func (m MsgCancelReplaceOrder) GetSignBytes() []byte {
	panic("implement me")
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
	panic("implement me")
}

func (m MsgCancelOrder) GetSignBytes() []byte {
	panic("implement me")
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
	panic("implement me")
}

func (m MsgAddOrder) GetSignBytes() []byte {
	panic("implement me")
}

func (m MsgAddOrder) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Owner}
}
