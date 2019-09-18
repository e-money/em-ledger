package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ sdk.Msg = MsgMintTokens{}
	_ sdk.Msg = MsgBurnTokens{}
)

type MsgMintTokens struct {
	Coins  sdk.Coins      `json:"coins"`
	Issuer sdk.AccAddress `json:"issuer"`
}

type MsgBurnTokens struct{}

func (m MsgMintTokens) Route() string {
	return "issuance"
}

func (m MsgMintTokens) Type() string {
	return "mint_tokens"
}

func (m MsgMintTokens) ValidateBasic() sdk.Error {
	fmt.Println(" *** Validating mint msg.")
	return nil
}

func (m MsgMintTokens) GetSignBytes() []byte {
	return []byte{}
}

func (m MsgMintTokens) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{}
}

func (m MsgBurnTokens) Route() string {
	return "issuance"
}

func (m MsgBurnTokens) Type() string {
	return "burn_tokens"
}

func (m MsgBurnTokens) ValidateBasic() sdk.Error {
	return nil
}

func (m MsgBurnTokens) GetSignBytes() []byte {
	return []byte{}
}

func (m MsgBurnTokens) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{}
}
