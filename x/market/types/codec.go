package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

var (
	amino = codec.NewLegacyAmino()

	// ModuleCdc references the module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	ModuleCdc = codec.NewAminoCodec(amino)
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgAddLimitOrder{}, "e-money/MsgAddLimitOrder", nil)
	cdc.RegisterConcrete(&MsgAddMarketOrder{}, "e-money/MsgAddMarketOrder", nil)
	cdc.RegisterConcrete(&MsgCancelReplaceLimitOrder{}, "e-money/MsgCancelReplaceLimitOrder", nil)
	cdc.RegisterConcrete(&MsgCancelOrder{}, "e-money/MsgCancelOrder", nil)
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgAddLimitOrder{},
		&MsgAddMarketOrder{},
		&MsgCancelReplaceLimitOrder{},
		&MsgCancelOrder{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	authtypes.RegisterLegacyAminoCodec(amino)
	banktypes.RegisterLegacyAminoCodec(amino)
	amino.Seal()
}
