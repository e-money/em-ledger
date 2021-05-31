// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"sync"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/e-money/em-ledger/util"
	"github.com/e-money/em-ledger/x/authority/types"
	"github.com/e-money/em-ledger/x/issuer"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	keyAuthorityAccAddress = "AuthorityAccountAddress"
	keyGasPrices           = "GasPrices"
)

type Keeper struct {
	cdc        codec.BinaryMarshaler
	storeKey   sdk.StoreKey
	ik         issuer.Keeper
	bankKeeper types.BankKeeper
	gpk        types.GasPricesKeeper

	gasPricesInit *sync.Once
}

func NewKeeper(cdc codec.BinaryMarshaler, storeKey sdk.StoreKey, issuerKeeper issuer.Keeper, bankKeeper types.BankKeeper, gasPricesKeeper types.GasPricesKeeper) Keeper {
	return Keeper{
		cdc:        cdc,
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

	bz := k.cdc.MustMarshalBinaryBare(&types.Authority{Address: authority.String()})
	store.Set([]byte(keyAuthorityAccAddress), bz)
}

func (k Keeper) GetAuthority(ctx sdk.Context) sdk.AccAddress {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(keyAuthorityAccAddress))
	var authority types.Authority
	k.cdc.MustUnmarshalBinaryBare(bz, &authority)
	acc, err := sdk.AccAddressFromBech32(authority.Address)
	if err != nil {
		panic(err.Error())
	}
	return acc
}

func (k Keeper) createIssuer(ctx sdk.Context, authority sdk.AccAddress, issuerAddress sdk.AccAddress, denoms []string) (*sdk.Result, error) {
	k.MustBeAuthority(ctx, authority)

	for _, denom := range denoms {
		if !util.ValidateDenom(denom) {
			return nil, sdkerrors.Wrapf(types.ErrInvalidDenom, "Invalid denom: %v", denom)
		}
	}

	i := issuer.NewIssuer(issuerAddress, denoms...)
	return k.ik.AddIssuer(ctx, i)
}

func (k Keeper) SetGasPrices(ctx sdk.Context, authority sdk.AccAddress, newPrices sdk.DecCoins) (*sdk.Result, error) {
	k.MustBeAuthority(ctx, authority)

	if !newPrices.IsValid() {
		return nil, sdkerrors.Wrapf(types.ErrInvalidGasPrices, "%v", newPrices)
	}

	// Check that the denominations actually exist before setting the gas prices to avoid being "locked out" of the blockchain
	supply := k.bankKeeper.GetSupply(ctx).GetTotal()
	for _, d := range newPrices {
		if supply.AmountOf(d.Denom).IsZero() {
			return nil, sdkerrors.Wrapf(types.ErrUnknownDenom, "%v", d.Denom)
		}
	}

	gasPrices := types.GasPrices{Minimum: newPrices}
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(&gasPrices)
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(keyGasPrices), bz)

	k.gpk.SetMinimumGasPrices(newPrices.String())
	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func (k Keeper) GetGasPrices(ctx sdk.Context) sdk.DecCoins {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(keyGasPrices))

	if bz == nil {
		return nil
	}
	var gasPrices types.GasPrices
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &gasPrices)
	return gasPrices.Minimum
}

func (k Keeper) destroyIssuer(ctx sdk.Context, authority sdk.AccAddress, issuerAddress sdk.AccAddress) (*sdk.Result, error) {
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

// Gas prices are kept in-memory in the app structure. Make sure they are initialized on node restart.
func (k Keeper) initGasPrices(ctx sdk.Context) {
	k.gasPricesInit.Do(func() {
		gps := k.GetGasPrices(ctx)
		k.gpk.SetMinimumGasPrices(gps.String())
	})
}
