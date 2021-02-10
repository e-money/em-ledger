// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package keeper

import (
	"sync"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/e-money/em-ledger/util"
	"github.com/e-money/em-ledger/x/authority/types"
	"github.com/e-money/em-ledger/x/issuer"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	keyAuthorityAccAddress = "AuthorityAccountAddress"
	keyRestrictedDenoms    = "RestrictedDenoms"
	keyGasPrices           = "GasPrices"
)

type Keeper struct {
	storeKey   sdk.StoreKey
	ik         issuer.Keeper
	bankKeeper types.BankKeeper
	gpk        types.GasPricesKeeper

	gasPricesInit *sync.Once
}

func NewKeeper(storeKey sdk.StoreKey, issuerKeeper issuer.Keeper, bankKeeper types.BankKeeper, gasPricesKeeper types.GasPricesKeeper) Keeper {
	return Keeper{
		ik:         issuerKeeper,
		bankKeeper: bankKeeper,
		gpk:        gasPricesKeeper,
		storeKey:   storeKey,

		gasPricesInit: new(sync.Once),
	}
}

func (k Keeper) SetAuthority(ctx sdk.Context, authority sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)

	if store.Get([]byte(keyAuthorityAccAddress)) != nil {
		panic("Authority was already specified")
	}

	bz := types.ModuleCdc.LegacyAmino.MustMarshalBinaryBare(authority)
	store.Set([]byte(keyAuthorityAccAddress), bz)
}

func (k Keeper) GetAuthority(ctx sdk.Context) (authority sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(keyAuthorityAccAddress))
	types.ModuleCdc.LegacyAmino.MustUnmarshalBinaryBare(bz, &authority)
	return
}

func (k Keeper) CreateIssuer(ctx sdk.Context, authority sdk.AccAddress, issuerAddress sdk.AccAddress, denoms []string) (*sdk.Result, error) {
	k.MustBeAuthority(ctx, authority)

	for _, denom := range denoms {
		if !util.ValidateDenom(denom) {
			return nil, sdkerrors.Wrapf(types.ErrInvalidDenom, "Invalid denom: %v", denom)
		}
	}

	i := issuer.NewIssuer(issuerAddress, denoms...)
	return k.ik.AddIssuer(ctx, i)
}

func (k Keeper) SetGasPrices(ctx sdk.Context, authority sdk.AccAddress, gasprices sdk.DecCoins) (*sdk.Result, error) {
	k.MustBeAuthority(ctx, authority)

	if !gasprices.IsValid() {
		return nil, sdkerrors.Wrapf(types.ErrInvalidGasPrices, "%v", gasprices)
	}

	// Check that the denominations actually exist before setting the gas prices to avoid being "locked out" of the blockchain
	supply := k.bankKeeper.GetSupply(ctx).GetTotal()
	for _, d := range gasprices {
		if supply.AmountOf(d.Denom).IsZero() {
			return nil, sdkerrors.Wrapf(types.ErrUnknownDenom, "%v", d.Denom)
		}
	}

	bz := types.ModuleCdc.LegacyAmino.MustMarshalBinaryLengthPrefixed(gasprices)
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(keyGasPrices), bz)

	k.gpk.SetMinimumGasPrices(gasprices.String())
	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func (k Keeper) GetGasPrices(ctx sdk.Context) (gasPrices sdk.DecCoins) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(keyGasPrices))

	if bz == nil {
		return
	}

	types.ModuleCdc.LegacyAmino.MustUnmarshalBinaryLengthPrefixed(bz, &gasPrices)
	return
}

func (k Keeper) DestroyIssuer(ctx sdk.Context, authority sdk.AccAddress, issuerAddress sdk.AccAddress) (*sdk.Result, error) {
	k.MustBeAuthority(ctx, authority)

	return k.ik.RemoveIssuer(ctx, issuerAddress)
}

func (k Keeper) MustBeAuthority(ctx sdk.Context, address sdk.AccAddress) {
	authority := k.GetAuthority(ctx)
	if authority == nil {
		panic(types.ErrNoAuthorityConfigured)
	}

	if authority.Equals(address) {
		return
	}

	panic(sdkerrors.Wrap(types.ErrNotAuthority, address.String()))
}

func (k Keeper) SetRestrictedDenoms(ctx sdk.Context, denoms types.RestrictedDenoms) {
	store := ctx.KVStore(k.storeKey)
	bz := types.ModuleCdc.LegacyAmino.MustMarshalBinaryLengthPrefixed(denoms)
	store.Set([]byte(keyRestrictedDenoms), bz)
}

func (k Keeper) GetRestrictedDenoms(ctx sdk.Context) (res types.RestrictedDenoms) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get([]byte(keyRestrictedDenoms))
	types.ModuleCdc.LegacyAmino.MustUnmarshalBinaryLengthPrefixed(bz, &res)

	return
}

// Gas prices are kept in-memory in the app structure. Make sure they are initialized on node restart.
func (k Keeper) initGasPrices(ctx sdk.Context) {
	k.gasPricesInit.Do(func() {
		gps := k.GetGasPrices(ctx)
		k.gpk.SetMinimumGasPrices(gps.String())
	})
}
