package keeper

import (
	"github.com/e-money/stargate/networks/emibctokens/x/swap/types"
)

var _ types.QueryServer = Keeper{}
