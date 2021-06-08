package ante

import sdk "github.com/cosmos/cosmos-sdk/types"

type StakingKeeper interface {
	BondDenom(sdk.Context) string
}
