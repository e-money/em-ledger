package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const RouterKey = ModuleName

var _ sdk.Msg = MsgSetInflation{}

type MsgSetInflation struct {
	Denom     string         `json:"denom"`
	Inflation sdk.Dec        `json:"inflation"`
	Principal sdk.AccAddress `json:"principal"` // The account that signed the request. Permission to set inflation is checked in handler.
}

func (msg MsgSetInflation) Route() string { return RouterKey }

func (msg MsgSetInflation) Type() string { return "set_inflation" }

func (msg MsgSetInflation) ValidateBasic() sdk.Error {
	if msg.Inflation.IsNegative() {
		return sdk.ErrUnknownRequest("Negative inflation is not supported")
	}

	return nil
}

func (msg MsgSetInflation) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg MsgSetInflation) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Principal}
}
