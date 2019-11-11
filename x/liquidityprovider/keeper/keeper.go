package keeper

import (
	"emoney/x/liquidityprovider/types"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/tendermint/tendermint/libs/log"
)

type Keeper struct {
	authKeeper   auth.AccountKeeper
	supplyKeeper supply.Keeper
}

func NewKeeper(ak auth.AccountKeeper, sk supply.Keeper) Keeper {
	return Keeper{
		authKeeper:   ak,
		supplyKeeper: sk,
	}
}

func (k Keeper) CreateLiquidityProvider(ctx sdk.Context, address sdk.AccAddress, mintable sdk.Coins) {
	logger := k.Logger(ctx)

	account := k.authKeeper.GetAccount(ctx, address)
	if account == nil {
		logger.Info("Account not found", "account", address)
		return
	}
	lpAcc := types.NewLiquidityProviderAccount(account, mintable)
	k.authKeeper.SetAccount(ctx, lpAcc)

	logger.Info("Created liquidity provider account.", "account", lpAcc.GetAddress())
}

func (k Keeper) BurnTokensFromBalance(ctx sdk.Context, liquidityProvider sdk.AccAddress, amount sdk.Coins) sdk.Result {
	account := k.GetLiquidityProviderAccount(ctx, liquidityProvider)
	if account == nil {
		return sdk.ErrUnknownAddress(fmt.Sprintf("account %s is not a liquidity provider or does not exist", liquidityProvider.String())).Result()
	}

	_, anynegative := account.Account.GetCoins().SafeSub(amount)
	if anynegative {
		return sdk.ErrInsufficientCoins(fmt.Sprintf("Insufficient balance for burn operation: %s < %s", account.Account.GetCoins(), amount)).Result()
	}

	err := k.supplyKeeper.SendCoinsFromAccountToModule(ctx, liquidityProvider, types.ModuleName, amount)
	if err != nil {
		return err.Result()
	}

	err = k.supplyKeeper.BurnCoins(ctx, types.ModuleName, amount)
	if err != nil {
		return err.Result()
	}

	account = k.GetLiquidityProviderAccount(ctx, liquidityProvider)
	account.Mintable = account.Mintable.Add(amount)
	k.SetLiquidityProviderAccount(ctx, account)

	return sdk.Result{Events: ctx.EventManager().Events()}
}

func (k Keeper) MintTokens(ctx sdk.Context, liquidityProvider sdk.AccAddress, amount sdk.Coins) sdk.Result {
	logger := k.Logger(ctx)

	account := k.GetLiquidityProviderAccount(ctx, liquidityProvider)
	if account == nil {
		return sdk.ErrUnknownAddress(fmt.Sprintf("account %s is not a liquidity provider or does not exist", liquidityProvider.String())).Result()
	}

	updatedMintableAmount, anyNegative := account.Mintable.SafeSub(amount)
	if anyNegative {
		logger.Debug(fmt.Sprintf("Insufficient mintable amount for minting operation"), "requested", amount, "available", account.Mintable)
		return sdk.ErrInsufficientCoins(fmt.Sprintf("insufficient liquidity provider mintable amount: %s < %s", account.Mintable, amount)).Result()
	}

	err := k.supplyKeeper.MintCoins(ctx, types.ModuleName, amount)
	if err != nil {
		return err.Result()
	}

	err = k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, liquidityProvider, amount)
	if err != nil {
		return err.Result()
	}

	account = k.GetLiquidityProviderAccount(ctx, liquidityProvider)
	account.Mintable = updatedMintableAmount
	k.SetLiquidityProviderAccount(ctx, account)

	return sdk.Result{Events: ctx.EventManager().Events()}
}

func (k Keeper) SetLiquidityProviderAccount(ctx sdk.Context, account *types.LiquidityProviderAccount) {
	k.authKeeper.SetAccount(ctx, account)
}

func (k Keeper) RevokeLiquidityProviderAccount(ctx sdk.Context, account auth.Account) bool {
	if lpAcc, isLpAcc := account.(*types.LiquidityProviderAccount); isLpAcc {
		account = lpAcc.Account
		k.authKeeper.SetAccount(ctx, account)
		return true
	}

	return false
}

func (k Keeper) GetLiquidityProviderAccount(ctx sdk.Context, liquidityProvider sdk.AccAddress) *types.LiquidityProviderAccount {
	logger := k.Logger(ctx)

	a := k.authKeeper.GetAccount(ctx, liquidityProvider)
	account, ok := a.(*types.LiquidityProviderAccount)
	if !ok {
		logger.Debug(fmt.Sprintf("Account is not a liquidity provider"), "address", liquidityProvider)
		return nil
	}

	return account
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
