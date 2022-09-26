// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"
)

var (
	_ sdk.Msg = &MsgCreateIssuer{}
	_ sdk.Msg = &MsgDestroyIssuer{}
	_ sdk.Msg = &MsgSetGasPrices{}
	_ sdk.Msg = &MsgReplaceAuthority{}
	_ sdk.Msg = &MsgScheduleUpgrade{}
	_ sdk.Msg = &MsgSetParameters{}
)

func (msg MsgDestroyIssuer) Type() string { return "destroy_issuer" }

func (msg MsgCreateIssuer) Type() string { return "create_issuer" }

func (msg MsgSetGasPrices) Type() string { return "set_gas_prices" }

func (msg MsgReplaceAuthority) Type() string { return "replace_authority" }

func (msg MsgScheduleUpgrade) Type() string { return "schedule_upgrade" }

func (msg MsgSetParameters) Type() string { return "set_parameters" }

func (msg MsgDestroyIssuer) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Issuer); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid issuer address (%s)", err)
	}
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid authority address (%s)", err)
	}

	return nil
}

func (msg MsgCreateIssuer) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Issuer); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid issuer address (%s)", err)
	}
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid authority address (%s)", err)
	}

	if len(msg.Denominations) == 0 {
		return sdkerrors.Wrap(ErrNoDenomsSpecified, "No denomination specified")
		// return ErrNoDenomsSpecified()
	}

	return nil
}

func (msg MsgSetGasPrices) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid authority address (%s)", err)
	}

	if !msg.GasPrices.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "%v", msg.GasPrices)
		// return sdk.ErrInvalidCoins(msg.GasPrices.String())
	}

	return nil
}

func (msg MsgReplaceAuthority) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid authority address (%s)", err)
	}

	if _, err := sdk.AccAddressFromBech32(msg.NewAuthority); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid new authority address (%s)", err)
	}

	return nil
}

func (msg MsgScheduleUpgrade) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid authority address (%s)", err)
	}

	if err := msg.Plan.ValidateBasic(); err != nil {
		return err
	}

	return nil
}

func (m MsgSetParameters) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid authority address (%s)", err)
	}

	if len(m.Changes) == 0 {
		return sdkerrors.Wrapf(ErrNoParams, "Message contains not parameter changes.")
	}

	if err := proposal.ValidateChanges(m.Changes); err != nil {
		return err
	}

	return nil
}

func (msg MsgDestroyIssuer) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg MsgCreateIssuer) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg MsgSetGasPrices) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg MsgReplaceAuthority) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg MsgScheduleUpgrade) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg MsgSetParameters) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg MsgDestroyIssuer) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgCreateIssuer) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgSetGasPrices) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgReplaceAuthority) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgScheduleUpgrade) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgSetParameters) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgDestroyIssuer) Route() string { return ModuleName }

func (msg MsgCreateIssuer) Route() string { return ModuleName }

func (msg MsgSetGasPrices) Route() string { return ModuleName }

func (msg MsgReplaceAuthority) Route() string { return ModuleName }

func (msg MsgScheduleUpgrade) Route() string { return ModuleName }

func (msg MsgSetParameters) Route() string { return ModuleName }
