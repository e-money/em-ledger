package keeper

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/inflation/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestQueryInflation(t *testing.T) {
	input := newTestInput(t)
	myState := types.NewInflationState(time.Now(), "ejpy", "0.05", "echf", "0.10", "eeur", "0.01")
	input.mintKeeper.SetState(input.ctx, myState)

	queryHelper := baseapp.NewQueryServerTestHelper(input.ctx, input.encConfig.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, input.mintKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	specs := map[string]struct {
		req      *types.QueryInflationRequest
		expErr   bool
		expState types.InflationState
	}{
		"all good": {
			req:      &types.QueryInflationRequest{},
			expState: myState,
		},
		"nil param": {
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			gotRsp, gotErr := queryClient.Inflation(sdk.WrapSDKContext(input.ctx), spec.req)
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, spec.expState, gotRsp.State)
		})
	}
}
