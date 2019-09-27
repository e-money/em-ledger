package types

import "github.com/cosmos/cosmos-sdk/codec"

var ModuleCdc *codec.Codec

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgCreateIssuer{}, "e-money/MsgCreateIssuer", nil)
	cdc.RegisterConcrete(MsgDestroyIssuer{}, "e-money/MsgDestroyIssuer", nil)
}

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
