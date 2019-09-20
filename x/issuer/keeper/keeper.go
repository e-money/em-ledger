package keeper

import (
	"emoney/x/issuer/types"
	lp "emoney/x/liquidityprovider"
	"fmt"
	"github.com/tendermint/tendermint/libs/log"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	keyIssuerList = "issuers"
)

type Keeper struct {
	storeKey sdk.StoreKey
	lpKeeper lp.Keeper
}

func NewKeeper(storeKey sdk.StoreKey, lpk lp.Keeper) Keeper {
	return Keeper{
		storeKey: storeKey,
		lpKeeper: lpk,
	}
}

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
		// Account was not previously a liquidity provider. Create it
		k.lpKeeper.CreateLiquidityProvider(ctx, liquidityProvider, creditIncrease)
	} else {
		log.Info("Increasing liquidity provider credit", "account", liquidityProvider, "increase", creditIncrease)
		lpAcc.IncreaseCredit(creditIncrease)
		k.lpKeeper.SetLiquidityProviderAccount(ctx, lpAcc)
	}
}

func (k Keeper) logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) getIssuers(ctx sdk.Context) (issuers []types.Issuer) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(keyIssuerList))
	types.ModuleCdc.MustUnmarshalBinaryBare(bz, &issuers)
	return
}

func (k Keeper) setIssuers(ctx sdk.Context, issuers []types.Issuer) {
	store := ctx.KVStore(k.storeKey)
	bz := types.ModuleCdc.MustMarshalBinaryBare(issuers)
	store.Set([]byte(keyIssuerList), bz)
}

func (k Keeper) AddIssuer(ctx sdk.Context, newIssuer types.Issuer) {
	issuers := k.getIssuers(ctx)

	existingDenoms := collectDenoms(issuers)
	if anyContained(existingDenoms, newIssuer.Denoms...) {
		panic(fmt.Errorf("denomination is already under control of an issuer"))
	}

	issuers = append(issuers, newIssuer)
	k.setIssuers(ctx, issuers)
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

	issuers := k.getIssuers(ctx)

	for _, issuer := range issuers {
		if issuer.Address.Equals(address) {
			return issuer
		}
	}

	k.logger(ctx).Info("Issuer operation attempted by non-issuer", "address", address)
	panic(fmt.Errorf("%v is not an issuer", address))
}
