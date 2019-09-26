package keeper

import (
	"fmt"

	"emoney/x/authority/types"
	"emoney/x/issuer"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	keyAuthorityAccAddress = "AuthorityAccountAddress"
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

func (k Keeper) setAuthority(ctx sdk.Context, authority sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)

	if store.Get([]byte(keyAuthorityAccAddress)) != nil {
		panic("Authority was already specified")
	}

	bz := types.ModuleCdc.MustMarshalBinaryBare(authority)
	store.Set([]byte(keyAuthorityAccAddress), bz)
}

func (k Keeper) getAuthority(ctx sdk.Context) (authority sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(keyAuthorityAccAddress))
	types.ModuleCdc.MustUnmarshalBinaryBare(bz, &authority)
	return
}

func (k Keeper) CreateIssuer(ctx sdk.Context, authority sdk.AccAddress, issuerAddress sdk.AccAddress, denoms []string) sdk.Error {
	k.MustBeAuthority(ctx, authority)

	for _, denom := range denoms {
		if !validateDenom(denom) {
			return types.ErrInvalidDenom(denom)
		}
	}

	i := issuer.NewIssuer(issuerAddress, denoms...)
	err := k.ik.AddIssuer(ctx, i)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) DestroyIssuer(ctx sdk.Context, authority sdk.AccAddress, issuerAddress sdk.AccAddress) sdk.Error {
	k.MustBeAuthority(ctx, authority)

	err := k.ik.RemoveIssuer(ctx, issuerAddress)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) MustBeAuthority(ctx sdk.Context, address sdk.AccAddress) {
	authority := k.getAuthority(ctx)
	if authority == nil {
		panic("Authority not set")
	}

	if authority.Equals(address) {
		return
	}

	panic(fmt.Errorf("address is not the authority: %v", address))
}

// The denomination validation functions are buried deep inside the Coin struct, so use this approach to validate names.
func validateDenom(denomination string) bool {
	defer func() {
		recover()
	}()
	// Function panics when encountering an invalid denomination
	sdk.NewCoin(denomination, sdk.ZeroInt())
	return true
}
