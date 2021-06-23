// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package keeper

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	"sort"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/e-money/em-ledger/x/issuer/types"
	lp "github.com/e-money/em-ledger/x/liquidityprovider"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"
)

const (
	keyIssuerList = "issuers"
)

type Keeper struct {
	cdc      codec.BinaryMarshaler
	storeKey sdk.StoreKey
	lpKeeper lp.Keeper
	ik       types.InflationKeeper
}

func NewKeeper(cdc codec.BinaryMarshaler, storeKey sdk.StoreKey, lpk lp.Keeper, ik types.InflationKeeper) Keeper {
	return Keeper{
		cdc:      cdc,
		storeKey: storeKey,
		lpKeeper: lpk,
		ik:       ik,
	}
}

func (k Keeper) IncreaseMintableAmountOfLiquidityProvider(ctx sdk.Context, liquidityProvider string, issuer sdk.AccAddress, mintableIncrease sdk.Coins) (*sdk.Result, error) {
	logger := k.logger(ctx)

	i, err := k.mustBeIssuer(ctx, issuer.String())
	if err != nil {
		return nil, err
	}

	for _, coin := range mintableIncrease {
		if !anyContained(i.Denoms, coin.Denom) {
			return nil, sdkerrors.Wrapf(types.ErrDoesNotControlDenomination, "%v", coin.Denom)
		}
	}

	lpAcc := k.lpKeeper.GetLiquidityProviderAccount(ctx, liquidityProvider)
	if lpAcc == nil {
		logger.Info("Creating liquidity provider", "account", liquidityProvider, "increase", mintableIncrease)
		// Account was not previously a liquidity provider. Create it
		if res, err := k.lpKeeper.CreateLiquidityProvider(ctx, liquidityProvider, mintableIncrease); err == nil {
			return res, nil
		}
	} else {
		logger.Info("Increasing liquidity provider mintable amount", "account", liquidityProvider, "increase", mintableIncrease)
		lpAcc.IncreaseMintableAmount(mintableIncrease)
		k.lpKeeper.SetLiquidityProviderAccount(ctx, lpAcc)
	}

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func (k Keeper) DecreaseMintableAmountOfLiquidityProvider(ctx sdk.Context, liquidityProvider string, issuer sdk.AccAddress, mintableDecrease sdk.Coins) (*sdk.Result, error) {
	logger := k.logger(ctx)

	i, err := k.mustBeIssuer(ctx, issuer.String())
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrNotAnIssuer, issuer.String())
	}

	for _, coin := range mintableDecrease {
		if !anyContained(i.Denoms, coin.Denom) {
			return nil, sdkerrors.Wrapf(types.ErrDoesNotControlDenomination, "%v", coin.Denom)
		}
	}

	lpAcc := k.lpKeeper.GetLiquidityProviderAccount(ctx, liquidityProvider)
	if lpAcc == nil {
		return nil, sdkerrors.Wrapf(types.ErrNotLiquidityProvider, "%v", liquidityProvider)
	}

	_, anyNegative := lpAcc.Mintable.SafeSub(mintableDecrease)
	if anyNegative {
		return nil, sdkerrors.Wrapf(types.ErrNegativeMintableBalance, "%v", lpAcc.String())
	}

	logger.Info("Liquidity provider mintable amount decreased", "account", liquidityProvider, "decrease", mintableDecrease)
	if err := lpAcc.DecreaseMintableAmount(mintableDecrease); err != nil {
		return nil, sdkerrors.Wrapf(types.ErrNegativeMintableBalance, "%v", lpAcc.String())
	}
	k.lpKeeper.SetLiquidityProviderAccount(ctx, lpAcc)

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil

}

func (k Keeper) RevokeLiquidityProvider(ctx sdk.Context, liquidityProvider string, issuerAddress sdk.AccAddress) (*sdk.Result, error) {
	issuer, err := k.mustBeIssuer(ctx, issuerAddress.String())
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrNotAnIssuer, issuerAddress.String())
	}

	lpAcc := k.lpKeeper.GetLiquidityProviderAccount(ctx, liquidityProvider)
	if lpAcc == nil {
		return nil, sdkerrors.Wrap(types.ErrNotLiquidityProvider, liquidityProvider)
	}

	newMintableAmount := lpAcc.Mintable
	for _, denom := range issuer.Denoms {
		newMintableAmount = removeDenom(newMintableAmount, denom)
	}

	if len(newMintableAmount) == len(lpAcc.Mintable) {
		// Nothing was changed. Issuer was not controlling this lp.
		return nil, sdkerrors.Wrap(types.ErrNotLiquidityProvider, liquidityProvider)
	}

	if len(newMintableAmount) == 0 {
		// Mintable amount is zero, so demote to ordinary account
		k.lpKeeper.RevokeLiquidityProviderAccount(ctx, lpAcc.Address)
	} else {
		// This liquidity provider has been granted a mintable amounts from multiple issuers so some amount remain.
		lpAcc.Mintable = newMintableAmount
		k.lpKeeper.SetLiquidityProviderAccount(ctx, lpAcc)
	}

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func (k Keeper) SetInflationRate(ctx sdk.Context, issuer sdk.AccAddress, inflationRate sdk.Dec, denom string) (*sdk.Result, error) {
	_, err := k.mustBeIssuerOfDenom(ctx, issuer.String(), denom)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrNotAnIssuer, issuer.String())
	}

	return k.ik.SetInflation(ctx, inflationRate, denom)
}

func (k Keeper) logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetIssuers(ctx sdk.Context) []types.Issuer {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(keyIssuerList))
	if bz == nil {
		return nil
	}

	var state types.Issuers
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &state)
	return state.Issuers
}

func (k Keeper) setIssuers(ctx sdk.Context, issuers []types.Issuer) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(&types.Issuers{Issuers: issuers})
	store.Set([]byte(keyIssuerList), bz)
}

func (k Keeper) AddIssuer(ctx sdk.Context, newIssuer types.Issuer) (*sdk.Result, error) {
	issuers := k.GetIssuers(ctx)

	existingDenoms := collectDenoms(issuers)
	if anyContained(existingDenoms, newIssuer.Denoms...) {
		return nil, sdkerrors.Wrapf(types.ErrDenominationAlreadyAssigned, "%v", newIssuer.Denoms)
	}

	found := false
	for i := range issuers {
		if issuers[i].Address == newIssuer.Address {
			issuers[i].Denoms = append(issuers[i].Denoms, newIssuer.Denoms...)
			sort.Strings(issuers[i].Denoms)
			found = true
			break
		}
	}

	if !found {
		issuers = append(issuers, newIssuer)
	}

	k.setIssuers(ctx, issuers)
	k.ik.AddDenoms(ctx, newIssuer.Denoms) // TODO Check error?
	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func (k Keeper) RemoveIssuer(ctx sdk.Context, issuer sdk.AccAddress) (*sdk.Result, error) {
	issuers := k.GetIssuers(ctx)

	updatedIssuers := make([]types.Issuer, 0)

	// This is one way to remove an element from a slice. There are many. This is one.
	for _, i := range issuers {
		if i.Address == issuer.String() {
			continue
		}

		updatedIssuers = append(updatedIssuers, i)
	}

	if len(updatedIssuers) == len(issuers) {
		return nil, sdkerrors.Wrap(types.ErrNotAnIssuer, issuer.String())
	}

	k.setIssuers(ctx, updatedIssuers)
	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func anyContained(s []string, searchterms ...string) bool {
	for _, st := range searchterms {
		index := sort.SearchStrings(s, st)
		if index < len(s) && s[index] == st {
			return true
		}
	}

	return false
}

func collectDenoms(issuers []types.Issuer) (res []string) {
	for _, issuer := range issuers {
		res = append(res, issuer.Denoms...)
	}

	sort.Strings(res)
	return
}

func removeDenom(coins sdk.Coins, denom string) (res sdk.Coins) {
	for _, c := range coins {
		if c.Denom == denom {
			continue
		}

		res = append(res, c)
	}

	return
}

func (k Keeper) mustBeIssuer(ctx sdk.Context, bech32Addr string) (types.Issuer, error) {
	if len(bech32Addr) == 0 {
		return types.Issuer{}, fmt.Errorf("no issuer specified")
	}

	issuers := k.GetIssuers(ctx)

	for _, issuer := range issuers {
		if issuer.Address == bech32Addr {
			return issuer, nil
		}
	}

	k.logger(ctx).Info("Issuer operation attempted by non-issuer", "address", bech32Addr)
	return types.Issuer{}, fmt.Errorf("%v is not an issuer", bech32Addr)
}

func (k Keeper) mustBeIssuerOfDenom(ctx sdk.Context, bech32Addr string, denom string) (types.Issuer, error) {
	if len(bech32Addr) == 0 {
		return types.Issuer{}, fmt.Errorf("no issuer specified")
	}

	issuers := k.GetIssuers(ctx)

	for _, issuer := range issuers {
		if issuer.Address == bech32Addr {
			for _, d := range issuer.Denoms {
				if d == denom {
					return issuer, nil
				}
			}

		}
	}

	k.logger(ctx).Info("Issuer operation attempted by non-issuer", "address", bech32Addr)
	return types.Issuer{}, fmt.Errorf("%v is not an issuer", bech32Addr)
}
