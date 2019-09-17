package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// generic sealed codec to be used throughout this module
var ModuleCdc *codec.Codec

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(LiquidityProviderAccount{}, "e-money/LiquidityProviderAccount", nil)
	cdc.RegisterConcrete(MsgMintTokens{}, "e-money/MsgMintTokens", nil)

	// TODO Remove
	cdc.RegisterConcrete(MsgDevTracerBullet{}, "e-money/MsgDevTracerBullet", nil)
}

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
