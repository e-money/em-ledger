// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package keeper

import (
	"fmt"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/inflation/internal/types"
)

type Keeper struct {
	cdc           *codec.LegacyAmino
	storeKey      sdk.StoreKey
	supplyKeeper  types.BankKeeper
	stakingKeeper types.StakingKeeper
	accountKeeper types.AccountKeeper

	cointokenDestination,
	stakingtokenDestination string
}

func NewKeeper(
	cdc *codec.LegacyAmino, key sdk.StoreKey, bankKeeper types.BankKeeper, accountKeeper types.AccountKeeper, stakingKeeper types.StakingKeeper, coinTokenDestination, stakingTokenDestination string) Keeper {

	if addr := accountKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic("the inflation module account has not been set")
	}

	return Keeper{
		cdc:           cdc,
		storeKey:      key,
		supplyKeeper:  bankKeeper,
		stakingKeeper: stakingKeeper,

		cointokenDestination:    coinTokenDestination,
		stakingtokenDestination: stakingTokenDestination,
	}
}

//______________________________________________________________________

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// get the minter
func (k Keeper) GetState(ctx sdk.Context) (is types.InflationState) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.MinterKey)
	if b == nil {
		panic("stored inflation state should not have been nil")
	}

	k.cdc.MustUnmarshalBinaryLengthPrefixed(b, &is)
	return
}

func (k Keeper) SetState(ctx sdk.Context, is types.InflationState) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryLengthPrefixed(is)
	store.Set(types.MinterKey, b)
}

func (k Keeper) SetInflation(ctx sdk.Context, newInflation sdk.Dec, denom string) (*sdk.Result, error) {
	state := k.GetState(ctx)
	asset := state.FindByDenom(denom)
	if asset == nil {
		return nil, sdkerrors.Wrapf(types.ErrUnknownRequest, "Unrecognized asset denomination: %v", denom)
	}

	asset.Inflation = newInflation
	k.SetState(ctx, state)

	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func (k Keeper) GetStakingDenomination(ctx sdk.Context) string {
	return k.stakingKeeper.GetParams(ctx).BondDenom
}

func (k Keeper) AddDenoms(ctx sdk.Context, denoms []string) (*sdk.Result, error) {
	state := k.GetState(ctx)

	for _, denom := range denoms {
		if state.FindByDenom(denom) != nil {
			continue
		}

		asset := types.InflationAsset{
			Denom:     denom,
			Inflation: sdk.ZeroDec(),
			Accum:     sdk.ZeroDec(),
		}

		state.InflationAssets = append(state.InflationAssets, asset)
	}

	k.SetState(ctx, state)
	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func (k Keeper) TotalTokenSupply(ctx sdk.Context) sdk.Coins {
	return k.supplyKeeper.GetSupply(ctx).GetTotal()
}

// MintCoins implements an alias call to the underlying supply keeper's
// MintCoins to be used in BeginBlocker.
func (k Keeper) MintCoins(ctx sdk.Context, newCoins sdk.Coins) error {
	if newCoins.Empty() {
		// skip as no coins need to be minted
		return nil
	}
	return k.supplyKeeper.MintCoins(ctx, types.ModuleName, newCoins)
}

func (k Keeper) DistributeMintedCoins(ctx sdk.Context, fees sdk.Coins) error {
	return k.supplyKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, k.cointokenDestination, fees)
}

func (k Keeper) DistributeStakingCoins(ctx sdk.Context, fees sdk.Coins) error {
	return k.supplyKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, k.stakingtokenDestination, fees)
}
