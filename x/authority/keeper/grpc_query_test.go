package keeper

import (
	"testing"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/authority/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryGasPrices(t *testing.T) {
	encConfig := MakeTestEncodingConfig()
	ctx, keeper, _, _ := createTestComponentWithEncodingConfig(t, encConfig)

	myGasPrices, _ := sdk.ParseDecCoins("0.400000000000000000echf,0.400000000000000000eeur")
	authority := mustParseAddress("emoney1kt0vh0ttget0xx77g6d3ttnvq2lnxx6vp3uyl0")
	keeper.BootstrapAuthority(ctx, authority)
	_, err := keeper.SetGasPrices(ctx, authority, myGasPrices)
	require.NoError(t, err)

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, encConfig.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, keeper)
	queryClient := types.NewQueryClient(queryHelper)

	specs := map[string]struct {
		req *types.QueryGasPricesRequest
	}{
		"all good": {
			req: &types.QueryGasPricesRequest{},
		},
		"nil param": {},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			gotRsp, gotErr := queryClient.GasPrices(sdk.WrapSDKContext(ctx), spec.req)
			require.NoError(t, gotErr)
			require.NotNil(t, gotRsp)
			assert.Equal(t, myGasPrices, gotRsp.MinGasPrices)
		})
	}
}

func TestQueryUpgradePlan(t *testing.T) {
	encConfig := MakeTestEncodingConfig()
	ctx, keeper, _, _ := createTestComponentWithEncodingConfig(t, encConfig)

	authority := mustParseAddress("emoney1kt0vh0ttget0xx77g6d3ttnvq2lnxx6vp3uyl0")
	keeper.BootstrapAuthority(ctx, authority)

	expPlan := upgradetypes.Plan{
		Name:   "expPlan 8",
		Height: 1000,
	}
	_, err := keeper.ScheduleUpgrade(ctx, authority, expPlan)
	require.NoError(t, err)

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, encConfig.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, keeper)
	queryClient := types.NewQueryClient(queryHelper)

	specs := map[string]struct {
		req *types.QueryUpgradePlanRequest
	}{
		"all good": {
			req: &types.QueryUpgradePlanRequest{},
		},
		"nil param": {},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			gotRsp, gotErr := queryClient.UpgradePlan(sdk.WrapSDKContext(ctx), spec.req)
			require.NoError(t, gotErr)
			require.NotNil(t, gotRsp)
			assert.Equal(t, expPlan, gotRsp.Plan)
		})
	}
}
