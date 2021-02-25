// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package keeper

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/e-money/em-ledger/x/liquidityprovider/types"
	"github.com/tendermint/tendermint/libs/log"
)

type Keeper struct {
	authKeeper types.AccountKeeper
	bankKeeper types.BankKeeper
}

func NewKeeper(ak types.AccountKeeper, bk types.BankKeeper) Keeper {
	return Keeper{
		authKeeper: ak,
		bankKeeper: bk,
	}
}

func (k Keeper) CreateLiquidityProvider(ctx sdk.Context, address sdk.AccAddress, mintable sdk.Coins) (*sdk.Result, error) {
	logger := k.Logger(ctx)

	account := k.authKeeper.GetAccount(ctx, address)
	if account == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownAddress, address.String())
	}
	lpAcc, err := types.NewLiquidityProviderAccount(account, mintable)
	if err != nil {
		return nil, err
	}
	k.authKeeper.SetAccount(ctx, lpAcc)

	logger.Info("Created liquidity provider account.", "account", lpAcc.GetAddress())
	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func (k Keeper) BurnTokensFromBalance(ctx sdk.Context, liquidityProvider sdk.AccAddress, amount sdk.Coins) (*sdk.Result, error) {
	account := k.GetLiquidityProviderAccount(ctx, liquidityProvider)
	if account == nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "account %s is not a liquidity provider or does not exist", liquidityProvider.String())
	}
	balances := k.bankKeeper.GetAllBalances(ctx, liquidityProvider)
	_, anynegative := balances.SafeSub(amount)
	if anynegative {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, "Insufficient balance for burn operation: %s < %s", balances, amount)
	}

	err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, liquidityProvider, types.ModuleName, amount)
	if err != nil {
		return nil, err
	}

	err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, amount)
	if err != nil {
		return nil, err
	}

	account.Mintable = account.Mintable.Add(amount...)
	k.SetLiquidityProviderAccount(ctx, account)

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func (k Keeper) MintTokens(ctx sdk.Context, liquidityProvider sdk.AccAddress, amount sdk.Coins) (*sdk.Result, error) {
	logger := k.Logger(ctx)

	account := k.GetLiquidityProviderAccount(ctx, liquidityProvider)
	if account == nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "account %s is not a liquidity provider or does not exist", liquidityProvider.String())
	}

	updatedMintableAmount, anyNegative := account.Mintable.SafeSub(amount)
	if anyNegative {
		logger.Debug(fmt.Sprintf("Insufficient mintable amount for minting operation"), "requested", amount, "available", account.Mintable)
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, "insufficient liquidity provider mintable amount: %s < %s", account.Mintable, amount)
	}

	err := k.bankKeeper.MintCoins(ctx, types.ModuleName, amount)
	if err != nil {
		return nil, err
	}

	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, liquidityProvider, amount)
	if err != nil {
		return nil, err
	}

	account = k.GetLiquidityProviderAccount(ctx, liquidityProvider)
	account.Mintable = updatedMintableAmount
	k.SetLiquidityProviderAccount(ctx, account)

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func (k Keeper) SetLiquidityProviderAccount(ctx sdk.Context, account *types.LiquidityProviderAccount) {
	k.authKeeper.SetAccount(ctx, account)
}

func (k Keeper) RevokeLiquidityProviderAccount(ctx sdk.Context, account authtypes.AccountI) bool {
	if lpAcc, isLpAcc := account.(*types.LiquidityProviderAccount); isLpAcc {
		account = lpAcc.GetNestedAccount()
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

func (k Keeper) GetAllLiquidityProviderAccounts(ctx sdk.Context) []types.LiquidityProviderAccount {
	res := make([]types.LiquidityProviderAccount, 0)

	k.authKeeper.IterateAccounts(ctx, func(acc authtypes.AccountI) (stop bool) {
		if lpAcc, ok := acc.(*types.LiquidityProviderAccount); ok {
			res = append(res, *lpAcc)
		}

		return false
	})

	return res
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
