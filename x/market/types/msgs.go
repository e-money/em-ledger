package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const ClientOrderIDMaxLength = 32

var (
	_ sdk.Msg = &MsgAddLimitOrder{}
	_ sdk.Msg = &MsgAddMarketOrder{}
	_ sdk.Msg = &MsgCancelOrder{}
	_ sdk.Msg = &MsgCancelReplaceLimitOrder{}
	_ sdk.Msg = &MsgCancelReplaceMarketOrder{}
)

func (m MsgAddMarketOrder) Route() string {
	return RouterKey
}

func (m MsgAddMarketOrder) Type() string {
	return "add_market_order"
}

func (m MsgAddMarketOrder) ValidateBasic() error {
	if m.MaxSlippage.LT(sdk.ZeroDec()) {
		return sdkerrors.Wrapf(ErrInvalidSlippage, "Cannot be negative")
	}

	if _, err := sdk.AccAddressFromBech32(m.Owner); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid owner address (%s)", err)
	}

	if !m.Destination.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "destination amount is invalid: %v", m.Destination.String())
	}

	err := sdk.ValidateDenom(m.Source)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "source denomination is invalid: %v", m.Source)
	}

	if m.Source == m.Destination.Denom {
		return sdkerrors.Wrapf(ErrInvalidInstrument, "'%v/%v' is not a valid instrument", m.Source, m.Destination.Denom)
	}

	return validateClientOrderID(m.ClientOrderId)

}

func (m MsgAddMarketOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

func (m MsgAddMarketOrder) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (m MsgCancelReplaceLimitOrder) Route() string {
	return RouterKey
}

func (m MsgCancelReplaceLimitOrder) Type() string {
	return "cancel_replace_limit_order"
}

func (m MsgCancelReplaceLimitOrder) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Owner); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid owner address (%s)", err)
	}

	if !m.Destination.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "destination amount is invalid: %v", m.Destination.String())
	}

	if !m.Source.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "source amount is invalid: %v", m.Source.String())
	}

	if m.Source.Denom == m.Destination.Denom {
		return sdkerrors.Wrapf(ErrInvalidInstrument, "'%v/%v' is not a valid instrument", m.Source.Denom, m.Destination.Denom)
	}

	err := validateClientOrderID(m.OrigClientOrderId)
	if err != nil {
		return err
	}

	return validateClientOrderID(m.NewClientOrderId)
}

func (m MsgCancelReplaceLimitOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

func (m MsgCancelReplaceLimitOrder) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (m MsgCancelOrder) Route() string {
	return RouterKey
}

func (m MsgCancelOrder) Type() string {
	return "cancel_order"
}

func (m MsgCancelOrder) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Owner); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid owner address (%s)", err)
	}

	return validateClientOrderID(m.ClientOrderId)
}

func (m MsgCancelOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

func (m MsgCancelOrder) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (m MsgAddLimitOrder) Route() string {
	return RouterKey
}

func (m MsgAddLimitOrder) Type() string {
	return "add_limit_order"
}

func (m MsgAddLimitOrder) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Owner); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid owner address (%s)", err)
	}

	if !m.Destination.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "destination amount is invalid: %v", m.Destination.String())
	}

	if !m.Source.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "source amount is invalid: %v", m.Source.String())
	}

	if m.Source.Denom == m.Destination.Denom {
		return sdkerrors.Wrapf(ErrInvalidInstrument, "'%v/%v' is not a valid instrument", m.Source.Denom, m.Destination.Denom)
	}

	return validateClientOrderID(m.ClientOrderId)
}

func (m MsgAddLimitOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

func (m MsgAddLimitOrder) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func validateClientOrderID(id string) error {
	if len(id) > ClientOrderIDMaxLength {
		return sdkerrors.Wrap(ErrInvalidClientOrderId, id)
	}

	return nil
}

func (m MsgCancelReplaceMarketOrder) Route() string {
	return RouterKey
}

func (m MsgCancelReplaceMarketOrder) Type() string {
	return "cancel_replace_market_order"
}

func (m MsgCancelReplaceMarketOrder) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Owner); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid owner address (%s)", err)
	}

	if !m.Destination.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "destination amount is invalid: %v", m.Destination.String())
	}

	if err := sdk.ValidateDenom(m.Source); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "source denomination is invalid: %v", m.Source)
	}

	if m.Source == m.Destination.Denom {
		return sdkerrors.Wrapf(ErrInvalidInstrument, "'%s/%s' is not a valid instrument", m.Source, m.Destination.Denom)
	}

	err := validateClientOrderID(m.OrigClientOrderId)
	if err != nil {
		return err
	}

	return validateClientOrderID(m.NewClientOrderId)
}

func (m MsgCancelReplaceMarketOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

func (m MsgCancelReplaceMarketOrder) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}
