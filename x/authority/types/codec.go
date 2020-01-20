// This software is Copyright (c) 2019 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import "github.com/cosmos/cosmos-sdk/codec"

var ModuleCdc *codec.Codec

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgCreateIssuer{}, "e-money/MsgCreateIssuer", nil)
	cdc.RegisterConcrete(MsgDestroyIssuer{}, "e-money/MsgDestroyIssuer", nil)
	cdc.RegisterConcrete(MsgSetGasPrices{}, "e-money/MsgSetGasPrices", nil)
}

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
