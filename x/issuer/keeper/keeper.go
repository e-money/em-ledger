package keeper

import (
	"fmt"
	"sort"

	"emoney/x/inflation"
	"emoney/x/issuer/types"
	lp "emoney/x/liquidityprovider"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"
)

const (
	keyIssuerList = "issuers"
)

type Keeper struct {
	storeKey sdk.StoreKey
	lpKeeper lp.Keeper
	ik       inflation.Keeper
}

func NewKeeper(storeKey sdk.StoreKey, lpk lp.Keeper, ik inflation.Keeper) Keeper {
	return Keeper{
		storeKey: storeKey,
		lpKeeper: lpk,
		ik:       ik,
	}
}

// TODO Return sdk.Results instead of panicking everywhere.
func (k Keeper) IncreaseCreditOfLiquidityProvider(ctx sdk.Context, liquidityProvider sdk.AccAddress, issuer sdk.AccAddress, creditIncrease sdk.Coins) {
	log := k.logger(ctx)

	i := k.mustBeIssuer(ctx, issuer)
	for _, coin := range creditIncrease {
		if !anyContained(i.Denoms, coin.Denom) {
			panic(fmt.Errorf("issuer %v cannot provide credit in %v", i.Address, coin.Denom))
		}
	}

	lpAcc := k.lpKeeper.GetLiquidityProviderAccount(ctx, liquidityProvider)
	if lpAcc == nil {
		log.Info("Creating liquidity provider", "account", liquidityProvider, "increase", creditIncrease)
		// Account was not previously a liquidity provider. Create it
		k.lpKeeper.CreateLiquidityProvider(ctx, liquidityProvider, creditIncrease)
	} else {
		log.Info("Increasing liquidity provider credit", "account", liquidityProvider, "increase", creditIncrease)
		lpAcc.IncreaseCredit(creditIncrease)
		k.lpKeeper.SetLiquidityProviderAccount(ctx, lpAcc)
	}
}

func (k Keeper) DecreaseCreditOfLiquidityProvider(ctx sdk.Context, liquidityProvider sdk.AccAddress, issuer sdk.AccAddress, creditDecrease sdk.Coins) sdk.Error {
	log := k.logger(ctx)

	i := k.mustBeIssuer(ctx, issuer)
	for _, coin := range creditDecrease {
		if !anyContained(i.Denoms, coin.Denom) {
			panic(fmt.Errorf("issuer %v cannot control credit in %v", i.Address, coin.Denom))
		}
	}

	lpAcc := k.lpKeeper.GetLiquidityProviderAccount(ctx, liquidityProvider)
	if lpAcc == nil {
		panic(fmt.Errorf("address %v is not a liquidity provider", liquidityProvider))
	}

	_, anyNegative := lpAcc.Credit.SafeSub(creditDecrease)
	if anyNegative {
		return types.ErrNegativeCredit(lpAcc.GetAddress())
	}

	log.Info("Liquidity provider credit decreased", "account", liquidityProvider, "decrease", creditDecrease)
	lpAcc.DecreaseCredit(creditDecrease)
	k.lpKeeper.SetLiquidityProviderAccount(ctx, lpAcc)

	return nil
}

func (k Keeper) RevokeLiquidityProvider(ctx sdk.Context, liquidityProvider sdk.AccAddress, issuer sdk.AccAddress) sdk.Error {
	k.mustBeIssuer(ctx, issuer)

	lpAcc := k.lpKeeper.GetLiquidityProviderAccount(ctx, liquidityProvider)
	if lpAcc == nil {
		return types.ErrNotLiquidityProvider(liquidityProvider)
	}

	if k.lpKeeper.RevokeLiquidityProviderAccount(ctx, lpAcc) {
		return nil
	}

	return types.ErrNotLiquidityProvider(liquidityProvider)
}

func (k Keeper) SetInflationRate(ctx sdk.Context, issuer sdk.AccAddress, inflationRate sdk.Dec, denom string) sdk.Error {
	k.mustBeIssuer(ctx, issuer)

	return k.ik.SetInflation(ctx, inflationRate, denom)
}

func (k Keeper) logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetIssuers(ctx sdk.Context) (issuers []types.Issuer) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(keyIssuerList))
	if bz == nil {
		return
	}

	types.ModuleCdc.MustUnmarshalBinaryLengthPrefixed(bz, &issuers)
	return
}

func (k Keeper) setIssuers(ctx sdk.Context, issuers []types.Issuer) {
	store := ctx.KVStore(k.storeKey)
	bz := types.ModuleCdc.MustMarshalBinaryLengthPrefixed(issuers)
	store.Set([]byte(keyIssuerList), bz)
}

func (k Keeper) AddIssuer(ctx sdk.Context, newIssuer types.Issuer) sdk.Error {
	issuers := k.GetIssuers(ctx)

	existingDenoms := collectDenoms(issuers)
	if anyContained(existingDenoms, newIssuer.Denoms...) {
		return types.ErrDenominationAlreadyAssigned()
	}

	issuers = append(issuers, newIssuer)
	k.setIssuers(ctx, issuers)
	return nil
}

func (k Keeper) RemoveIssuer(ctx sdk.Context, issuer sdk.AccAddress) sdk.Error {
	issuers := k.GetIssuers(ctx)

	updatedIssuers := make([]types.Issuer, 0)

	// This is one way to remove an element from a slice. There are many. This is one.
	for _, i := range issuers {
		if i.Address.Equals(issuer) {
			continue
		}

		updatedIssuers = append(updatedIssuers, i)
	}

	if len(updatedIssuers) == len(issuers) {
		return types.ErrIssuerNotFound(issuer)
	}

	k.setIssuers(ctx, updatedIssuers)
	return nil
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

func (k Keeper) mustBeIssuer(ctx sdk.Context, address sdk.AccAddress) types.Issuer {
	if address == nil {
		panic(fmt.Errorf("%v is not an issuer", address))
	}

	issuers := k.GetIssuers(ctx)

	for _, issuer := range issuers {
		if issuer.Address.Equals(address) {
			return issuer
		}
	}

	k.logger(ctx).Info("Issuer operation attempted by non-issuer", "address", address)
	panic(fmt.Errorf("%v is not an issuer", address))
}
