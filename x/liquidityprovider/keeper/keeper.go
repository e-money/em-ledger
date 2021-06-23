// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package keeper

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/e-money/em-ledger/x/liquidityprovider/types"
	"github.com/tendermint/tendermint/libs/log"
	"sort"
)

type Keeper struct {
	cdc        codec.BinaryMarshaler
	storeKey   sdk.StoreKey
	bankKeeper types.BankKeeper
}

func NewKeeper(
	cdc codec.BinaryMarshaler, key sdk.StoreKey, bk types.BankKeeper,
) Keeper {
	return Keeper{
		cdc:        cdc,
		storeKey:   key,
		bankKeeper: bk,
	}
}

// ------------------------------------------
//				State functions
// ------------------------------------------

// SetLiquidityProviderAccount stores a liquidity provider in the state.
func (k Keeper) SetLiquidityProviderAccount(
	ctx sdk.Context, prov *types.LiquidityProviderAccount,
) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.ProviderKeyPrefix)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(prov)
	store.Set([]byte(prov.Address), bz)
}

// GetLiquidityProviderAccount finds and deserializes a liquidity provider from the state.
func (k Keeper) GetLiquidityProviderAccount(ctx sdk.Context, address sdk.AccAddress) *types.LiquidityProviderAccount {
	var prov types.LiquidityProviderAccount

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.ProviderKeyPrefix)

	bz := store.Get([]byte(address.String()))
	if bz == nil {
		return nil
	}

	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &prov)
	return &prov
}

func (k Keeper) RevokeLiquidityProviderAccount(ctx sdk.Context, address sdk.AccAddress) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.ProviderKeyPrefix)
	store.Delete([]byte(address.String()))
}

// IterateProviders provides an iterator over all stored liquidity providers.
// For each liquidity provider, cb will be called. If cb returns true, the
// iterator will close and stop.
func (k Keeper) IterateProviders(
	ctx sdk.Context,
	cb func(prov types.LiquidityProviderAccount) (stop bool),
) {
	iterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.ProviderKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var prov types.LiquidityProviderAccount
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &prov)

		if cb(prov) {
			break
		}
	}
}

// GetAllLiquidityProviderAccounts returns all the valid liquidity providers.
func (k Keeper) GetAllLiquidityProviderAccounts(ctx sdk.Context) []types.LiquidityProviderAccount {
	res := make([]types.LiquidityProviderAccount, 0)

	k.IterateProviders(
		ctx, func(prov types.LiquidityProviderAccount) (stop bool) {
			if err := prov.Validate(); err == nil {
				res = append(res, prov)
			}

			return false
		},
	)

	sort.Slice(res, func(i, j int) bool {
		return res[i].Address < res[j].Address
	})

	return res
}

func (k Keeper) CreateLiquidityProvider(ctx sdk.Context, address sdk.AccAddress, mintable sdk.Coins) (*sdk.Result, error) {
	logger := k.Logger(ctx)

	lpAcc, err := types.NewLiquidityProviderAccount(address.String(), mintable)
	if err != nil {
		return nil, err
	}
	k.SetLiquidityProviderAccount(ctx, lpAcc)

	logger.Info("Created liquidity provider account.", "account", lpAcc.Address)
	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

// ------------------------------------------
//				Banking functions
// ------------------------------------------

func (k Keeper) BurnTokensFromBalance(ctx sdk.Context, liquidityProvider sdk.AccAddress, amount sdk.Coins) (*sdk.Result, error) {
	prov := k.GetLiquidityProviderAccount(ctx, liquidityProvider)
	if prov == nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress,
			"account %s is not a liquidity provider or does not exist", liquidityProvider.String(),
		)
	}

	balances := k.bankKeeper.GetAllBalances(ctx, liquidityProvider)
	_, anynegative := balances.SafeSub(amount)
	if anynegative {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrInsufficientFunds,
			"Insufficient balance for burn operation: %s < %s", balances, amount,
		)
	}

	err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, liquidityProvider, types.ModuleName, amount)
	if err != nil {
		return nil, err
	}

	err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, amount)
	if err != nil {
		return nil, err
	}

	prov.Mintable = prov.Mintable.Add(amount...)
	k.SetLiquidityProviderAccount(ctx, prov)

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func (k Keeper) MintTokens(ctx sdk.Context, liquidityProvider sdk.AccAddress, amount sdk.Coins) (*sdk.Result, error) {
	logger := k.Logger(ctx)

	prov := k.GetLiquidityProviderAccount(ctx, liquidityProvider)
	if prov == nil {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrUnknownAddress,
			"account %s is not a liquidity provider or does not exist",
			liquidityProvider,
		)
	}

	updatedMintableAmount, anyNegative := prov.Mintable.SafeSub(amount)
	if anyNegative {
		logger.Debug(
			fmt.Sprintf("Insufficient mintable amount for minting operation"),
			"requested", amount, "available", prov.Mintable,
		)
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrInsufficientFunds,
			"insufficient liquidity provider mintable amount: %s < %s",
			prov.Mintable, amount,
		)
	}

	err := k.bankKeeper.MintCoins(ctx, types.ModuleName, amount)
	if err != nil {
		return nil, err
	}

	err = k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx, types.ModuleName, liquidityProvider, amount,
	)
	if err != nil {
		return nil, err
	}

	prov.Mintable = updatedMintableAmount
	k.SetLiquidityProviderAccount(ctx, prov)

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
