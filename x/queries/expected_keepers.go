package queries

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type AccountKeeper interface {
	GetModuleAccount(ctx sdk.Context, moduleName string) authtypes.ModuleAccountI
}

type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
<<<<<<< HEAD
	GetSupply(ctx sdk.Context, denom string) sdk.Coin
	IterateAllBalances(ctx sdk.Context, cb func(sdk.AccAddress, sdk.Coin) bool)
	IterateAllDenomMetaData(ctx sdk.Context, cb func(banktypes.Metadata) bool)
	GetAllDenomMetaData(ctx sdk.Context) []banktypes.Metadata
=======
	GetSupply(ctx sdk.Context) exported.SupplyI
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
>>>>>>> 9603b15 (Refactor the calculation of circulating supply to avoid iterating all accounts)
}

type SlashingKeeper interface {
	GetMissedBlocks(ctx sdk.Context, consAddr sdk.ConsAddress) (int64, int64)
}
