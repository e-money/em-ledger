package v09

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v039"
)

type LiquidityProviderAccount struct {
	auth.BaseAccount

	Mintable sdk.Coins `json:"mintable" yaml:"mintable"`
}

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&LiquidityProviderAccount{}, "e-money/LiquidityProviderAccount", nil)
	//cdc.RegisterConcrete(MsgMintTokens{}, "e-money/MsgMintTokens", nil)
	//cdc.RegisterConcrete(MsgBurnTokens{}, "e-money/MsgBurnTokens", nil)

	//cdc.RegisterInterface((*v038auth.GenesisAccount)(nil), nil)
	//cdc.RegisterInterface((*v038auth.Account)(nil), nil)
	//cdc.RegisterConcrete(&BaseAccount{}, "cosmos-sdk/Account", nil)
	//cdc.RegisterConcrete(&BaseVestingAccount{}, "cosmos-sdk/BaseVestingAccount", nil)
	//cdc.RegisterConcrete(&ContinuousVestingAccount{}, "cosmos-sdk/ContinuousVestingAccount", nil)
	//cdc.RegisterConcrete(&DelayedVestingAccount{}, "cosmos-sdk/DelayedVestingAccount", nil)
	//cdc.RegisterConcrete(&PeriodicVestingAccount{}, "cosmos-sdk/PeriodicVestingAccount", nil)
	//cdc.RegisterConcrete(&ModuleAccount{}, "cosmos-sdk/ModuleAccount", nil)
}
