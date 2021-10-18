package ante

import sdk "github.com/cosmos/cosmos-sdk/types"

type StakingKeeper interface {
	BondDenom(sdk.Context) string
}

// FeegrantKeeper defines the expected feegrant keeper.
type FeegrantKeeper interface {
	UseGrantedFees(ctx sdk.Context, granter, grantee sdk.AccAddress, fee sdk.Coins, msgs []sdk.Msg) error
}
