// This software is Copyright (c) 2019 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package keeper

import (
	emtypes "github.com/e-money/em-ledger/types"
	"github.com/e-money/em-ledger/util"
	"github.com/e-money/em-ledger/x/authority/types"
	"github.com/e-money/em-ledger/x/issuer"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	keyAuthorityAccAddress = "AuthorityAccountAddress"
	keyRestrictedDenoms    = "RestrictedDenoms"
)

type Keeper struct {
	storeKey sdk.StoreKey
	ik       issuer.Keeper
}

func NewKeeper(storeKey sdk.StoreKey, issuerKeeper issuer.Keeper) Keeper {
	return Keeper{
		ik:       issuerKeeper,
		storeKey: storeKey,
	}
}

func (k Keeper) SetAuthority(ctx sdk.Context, authority sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)

	if store.Get([]byte(keyAuthorityAccAddress)) != nil {
		panic("Authority was already specified")
	}

	bz := types.ModuleCdc.MustMarshalBinaryBare(authority)
	store.Set([]byte(keyAuthorityAccAddress), bz)
}

func (k Keeper) GetAuthority(ctx sdk.Context) (authority sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(keyAuthorityAccAddress))
	types.ModuleCdc.MustUnmarshalBinaryBare(bz, &authority)
	return
}

func (k Keeper) CreateIssuer(ctx sdk.Context, authority sdk.AccAddress, issuerAddress sdk.AccAddress, denoms []string) sdk.Result {
	k.MustBeAuthority(ctx, authority)

	for _, denom := range denoms {
		if !util.ValidateDenom(denom) {
			return types.ErrInvalidDenom(denom).Result()
		}
	}

	i := issuer.NewIssuer(issuerAddress, denoms...)
	return k.ik.AddIssuer(ctx, i)
}

func (k Keeper) DestroyIssuer(ctx sdk.Context, authority sdk.AccAddress, issuerAddress sdk.AccAddress) sdk.Result {
	k.MustBeAuthority(ctx, authority)

	return k.ik.RemoveIssuer(ctx, issuerAddress)
}

func (k Keeper) MustBeAuthority(ctx sdk.Context, address sdk.AccAddress) {
	authority := k.GetAuthority(ctx)
	if authority == nil {
		panic(types.ErrNoAuthorityConfigured())
	}

	if authority.Equals(address) {
		return
	}

	panic(types.ErrNotAuthority(address.String()))
}

func (k Keeper) SetRestrictedDenoms(ctx sdk.Context, denoms emtypes.RestrictedDenoms) {
	store := ctx.KVStore(k.storeKey)
	bz := types.ModuleCdc.MustMarshalBinaryLengthPrefixed(denoms)
	store.Set([]byte(keyRestrictedDenoms), bz)
}

func (k Keeper) GetRestrictedDenoms(ctx sdk.Context) (res emtypes.RestrictedDenoms) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get([]byte(keyRestrictedDenoms))
	types.ModuleCdc.MustUnmarshalBinaryLengthPrefixed(bz, &res)

	return
}
