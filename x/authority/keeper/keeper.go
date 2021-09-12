// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package keeper

import (
	"errors"
	"sync"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/e-money/em-ledger/x/authority/types"
	"github.com/e-money/em-ledger/x/issuer"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	keyAuthorityAccAddress = "AuthorityAccountAddress"
	keyGasPrices           = "GasPrices"
)

var _ authorityKeeper = Keeper{}

type Keeper struct {
	cdc           codec.BinaryMarshaler
	storeKey      sdk.StoreKey
	ik            issuer.Keeper
	bankKeeper    types.BankKeeper
	upgradeKeeper types.UpgradeKeeper
	gpk           types.GasPricesKeeper

	gasPricesInit *sync.Once
}

func NewKeeper(
	cdc codec.BinaryMarshaler, storeKey sdk.StoreKey,
	issuerKeeper issuer.Keeper, bankKeeper types.BankKeeper,
	gasPricesKeeper types.GasPricesKeeper, upgradeKeeper types.UpgradeKeeper,
) Keeper {
	return Keeper{
		cdc:           cdc,
		ik:            issuerKeeper,
		bankKeeper:    bankKeeper,
		gpk:           gasPricesKeeper,
		storeKey:      storeKey,
		upgradeKeeper: upgradeKeeper,

		gasPricesInit: new(sync.Once),
	}
}

// BootstrapAuthority solely exists for the genesis establishment of the chain
// authority. Once the authority is set, invoking this function again will panic.
func (k Keeper) BootstrapAuthority(ctx sdk.Context, newAuthority sdk.AccAddress) {
	authorityAcc, _, _ := k.getAuthorities(ctx)

	// set authority only if it is not set up.
	if !authorityAcc.Empty() {
		panic(errors.New("authority is set and sealed"))
	}

	k.saveAuthorities(ctx, newAuthority, "")
}

func (k Keeper) saveAuthorities(
	ctx sdk.Context, newAuthority sdk.AccAddress, formerAuthorityAddr string,
) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshalBinaryBare(
		&types.Authority{
			Address:       newAuthority.String(),
			FormerAddress: formerAuthorityAddr,
			LastModified:  ctx.BlockTime(),
		},
	)
	store.Set([]byte(keyAuthorityAccAddress), bz)
}

func (k Keeper) getAuthorities(ctx sdk.Context) (authority, formerAuthority sdk.AccAddress, err error) {
	authoritySet := k.GetAuthoritySet(ctx)

	authority, err = sdk.AccAddressFromBech32(authoritySet.Address)
	if err != nil {
		return nil, nil, err
	}

	withinTransitionPeriod := authoritySet.LastModified.
		Add(types.AuthorityTransitionDuration).
		After(ctx.BlockTime())
	if withinTransitionPeriod && authoritySet.FormerAddress != "" {
		// we keep the former address
		formerAuthority, _ = sdk.AccAddressFromBech32(authoritySet.FormerAddress)
	}

	return authority, formerAuthority, nil
}

func (k Keeper) GetAuthoritySet(ctx sdk.Context) types.Authority {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(keyAuthorityAccAddress))
	var authoritySet types.Authority
	k.cdc.MustUnmarshalBinaryBare(bz, &authoritySet)
	return authoritySet
}

func (k Keeper) createIssuer(ctx sdk.Context, authority sdk.AccAddress, issuerAddress sdk.AccAddress, denomsMetaData []types.Denomination) (*sdk.Result, error) {
	if err := k.ValidateAuthority(ctx, authority); err != nil {
		return nil, err
	}

	denoms := make([]string, len(denomsMetaData))
	for i, denomMetadatum := range denomsMetaData {
		if err := sdk.ValidateDenom(denomMetadatum.Base); err != nil {
			return nil, sdkerrors.Wrapf(types.ErrInvalidDenom, err.Error())
		}
		denoms[i] = denomMetadatum.Base
	}

	i := issuer.NewIssuer(issuerAddress, denoms...)
	return k.ik.AddIssuer(ctx, i, denomsMetaData)
}

func (k Keeper) SetGasPrices(ctx sdk.Context, authority sdk.AccAddress, newPrices sdk.DecCoins) (*sdk.Result, error) {
	if err := k.ValidateAuthority(ctx, authority); err != nil {
		return nil, err
	}

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

	if err := k.gpk.SetMinimumGasPrices(newPrices.String()); err != nil {
		return nil, err
	}

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
	if err := k.ValidateAuthority(ctx, authority); err != nil {
		return nil, err
	}

	return k.ik.RemoveIssuer(ctx, issuerAddress)
}

func (k Keeper) ValidateAuthority(ctx sdk.Context, address sdk.AccAddress) error {
	authority, formerAuth, err := k.getAuthorities(ctx)
	if err != nil {
		return sdkerrors.Wrap(types.ErrNoAuthorityConfigured, err.Error())
	}

	if authority == nil {
		return sdkerrors.Wrap(types.ErrNoAuthorityConfigured, address.String())
	}

	if formerAuth == nil {
		formerAuth = authority
	}

	if !authority.Equals(address) && !formerAuth.Equals(address) {
		return sdkerrors.Wrap(types.ErrNotAuthority, address.String())
	}

	return nil
}

// Gas prices are kept in-memory in the app structure. Make sure they are initialized on node restart.
func (k Keeper) initGasPrices(ctx sdk.Context) {
	k.gasPricesInit.Do(func() {
		gps := k.GetGasPrices(ctx)
		k.gpk.SetMinimumGasPrices(gps.String())
	})
}

func (k Keeper) replaceAuthority(ctx sdk.Context, authority, newAuthority sdk.AccAddress) (*sdk.Result, error) {
	if err := k.ValidateAuthority(ctx, authority); err != nil {
		return nil, err
	}

	_, formerAuthorityAddress, err := k.getAuthorities(ctx)
	if err != nil || formerAuthorityAddress.Empty() {
		formerAuthorityAddress = authority
	}

	k.saveAuthorities(ctx, newAuthority, formerAuthorityAddress.String())

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func (k Keeper) ScheduleUpgrade(
	ctx sdk.Context, authority sdk.AccAddress, plan upgradetypes.Plan,
) (*sdk.Result, error) {
	if err := k.ValidateAuthority(ctx, authority); err != nil {
		return nil, err
	}

	if err := k.upgradeKeeper.ScheduleUpgrade(ctx, plan); err != nil {
		return nil, err
	}

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func (k Keeper) GetUpgradePlan(ctx sdk.Context) (plan upgradetypes.Plan, havePlan bool) {

	return k.upgradeKeeper.GetUpgradePlan(ctx)
}
