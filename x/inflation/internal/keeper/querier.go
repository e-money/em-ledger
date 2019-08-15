package keeper

import (
	"encoding/json"
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	"emoney/x/inflation/internal/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewQuerier returns a minting Querier handler.
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, _ abci.RequestQuery) ([]byte, sdk.Error) {
		switch path[0] {

		case types.QueryInflation:
			return queryInflation(ctx, k)

		default:
			return nil, sdk.ErrUnknownRequest(fmt.Sprintf("unknown minting query endpoint: %s", path[0]))
		}
	}
}

func queryInflation(ctx sdk.Context, k Keeper) ([]byte, sdk.Error) {
	inflationState := k.GetState(ctx)

	// TODO Introduce a more presentation-friendly response type
	res, err := json.Marshal(inflationState)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to marshal JSON", err.Error()))
	}

	return res, nil
}
