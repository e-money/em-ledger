package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

var (
	amino = codec.NewLegacyAmino()

	// ModuleCdc references the module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	ModuleCdc = codec.NewAminoCodec(amino)
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgIncreaseMintable{}, "e-money/MsgIncreaseMintable", nil)
	cdc.RegisterConcrete(&MsgDecreaseMintable{}, "e-money/MsgDecreaseMintable", nil)
	cdc.RegisterConcrete(&MsgRevokeLiquidityProvider{}, "e-money/MsgRevokeLiquidityProvider", nil)
	cdc.RegisterConcrete(&MsgSetInflation{}, "e-money/MsgSetInflation", nil)
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgIncreaseMintable{},
		&MsgDecreaseMintable{},
		&MsgRevokeLiquidityProvider{},
		&MsgSetInflation{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}
