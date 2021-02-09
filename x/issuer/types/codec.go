// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
}

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}
