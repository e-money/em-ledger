package swap

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/stargate/networks/emibctokens/x/swap/keeper"
	"github.com/e-money/stargate/networks/emibctokens/x/swap/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// this line is used by starport scaffolding # genesis/module/init
	// Set all the denomTrace
	for _, elem := range genState.DenomTraceList {
		k.SetDenomTrace(ctx, *elem)
	}

	// Set all the ibcToken
	for _, elem := range genState.IbcTokenList {
		k.SetIbcToken(ctx, *elem)
	}

	k.SetPort(ctx, genState.PortId)
	// Only try to bind to port if it is not already bound, since we may already own
	// port capability from capability InitGenesis
	if !k.IsBound(ctx, genState.PortId) {
		// module binds to the port on InitChain
		// and claims the returned capability
		err := k.BindPort(ctx, genState.PortId)
		if err != nil {
			panic("could not claim port capability: " + err.Error())
		}
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()

	// this line is used by starport scaffolding # genesis/module/export
	// Get all denomTrace
	denomTraceList := k.GetAllDenomTrace(ctx)
	for _, elem := range denomTraceList {
		elem := elem
		genesis.DenomTraceList = append(genesis.DenomTraceList, &elem)
	}

	// Get all ibcToken
	ibcTokenList := k.GetAllIbcToken(ctx)
	for _, elem := range ibcTokenList {
		elem := elem
		genesis.IbcTokenList = append(genesis.IbcTokenList, &elem)
	}

	genesis.PortId = k.GetPort(ctx)

	return genesis
}
