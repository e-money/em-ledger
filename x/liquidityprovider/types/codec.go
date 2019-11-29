// This software is Copyright (c) 2019 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// generic sealed codec to be used throughout this module
var ModuleCdc *codec.Codec

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(&LiquidityProviderAccount{}, "e-money/LiquidityProviderAccount", nil)
	cdc.RegisterConcrete(MsgMintTokens{}, "e-money/MsgMintTokens", nil)
	cdc.RegisterConcrete(MsgBurnTokens{}, "e-money/MsgBurnTokens", nil)
}

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
