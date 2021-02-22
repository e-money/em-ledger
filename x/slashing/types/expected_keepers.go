package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankexported "github.com/cosmos/cosmos-sdk/x/bank/exported"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
)

type BankKeeper interface {
	GetSupply(ctx sdk.Context) bankexported.SupplyI
	SendCoinsFromModuleToModule(ctx sdk.Context, senderPool, recipientPool string, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
}

type StakingKeeper interface {
	slashingtypes.StakingKeeper
	BondDenom(ctx sdk.Context) string
}

// ParamSubspace defines the expected Subspace interfacace
type ParamSubspace interface {
	HasKeyTable() bool
	WithKeyTable(table paramtypes.KeyTable) paramtypes.Subspace
	Get(ctx sdk.Context, key []byte, ptr interface{})
	GetParamSet(ctx sdk.Context, ps paramtypes.ParamSet)
	SetParamSet(ctx sdk.Context, ps paramtypes.ParamSet)
}
