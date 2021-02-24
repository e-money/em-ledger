package keeper

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/authority/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestQueryGasPrices(t *testing.T) {
	encConfig := MakeTestEncodingConfig()
	ctx, keeper, _, _ := createTestComponentWithEncodingConfig(t, encConfig)

	myGasPrices, _ := sdk.ParseDecCoins("0.400000000000000000echf,0.400000000000000000eeur")
	authority := mustParseAddress("emoney1kt0vh0ttget0xx77g6d3ttnvq2lnxx6vp3uyl0")
	keeper.SetAuthority(ctx, authority)
	_, err := keeper.SetGasPrices(ctx, authority, myGasPrices)
	require.NoError(t, err)

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, encConfig.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, keeper)
	queryClient := types.NewQueryClient(queryHelper)

	specs := map[string]struct {
		req    *types.QueryGasPricesRequest
		expErr bool
	}{
		"all good": {
			req: &types.QueryGasPricesRequest{},
		},
		"nil param": {
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			gotRsp, gotErr := queryClient.GasPrices(sdk.WrapSDKContext(ctx), spec.req)
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, myGasPrices, gotRsp.MinGasPrices)
		})
	}
}
