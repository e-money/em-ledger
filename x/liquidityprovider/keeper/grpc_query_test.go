package keeper

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/e-money/em-ledger/x/liquidityprovider/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/rand"
	"sort"
	"testing"
)

func TestQueryList(t *testing.T) {
	var req = &types.QueryListRequest{}

	encConfig := MakeTestEncodingConfig()
	initialSupply := sdk.NewCoins(
		sdk.NewCoin("eeur", sdk.NewIntWithDecimal(50, 2)),
		sdk.NewCoin("ejpy", sdk.NewInt(250)),
	)
	ctx, _, _, keeper := createTestComponents(t, initialSupply)

	accAddr1 := sdk.AccAddress(rand.Bytes(sdk.AddrLen))
	addr1 := accAddr1.String()

	accAddr2 := sdk.AccAddress(rand.Bytes(sdk.AddrLen))
	addr2 := accAddr2.String()

	mintable1 := sdk.NewCoins(
		sdk.NewCoin("eeur", sdk.NewIntWithDecimal(1000, 2)),
	)

	mintable2 := sdk.NewCoins(
		sdk.NewCoin("ejpy", sdk.NewIntWithDecimal(888, 2)),
	)

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, encConfig.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, keeper)
	queryClient := types.NewQueryClient(queryHelper)

	specs := map[string]struct {
		setupFunc func()
		expLps    []types.LiquidityProviderAccount
		expErr    bool
	}{
		"all good": {
			setupFunc: func() {
				_, err := keeper.CreateLiquidityProvider(ctx,accAddr1, mintable1)
				assert.NoError(t, err)

				lp1 := keeper.GetLiquidityProviderAccount(ctx,accAddr1)
				assert.NotNil(t, lp1)

				_, err = keeper.CreateLiquidityProvider(ctx,accAddr2, mintable2)
				assert.NoError(t, err)

				lp2 := keeper.GetLiquidityProviderAccount(ctx,accAddr2)
				assert.NotNil(t, lp2)
			},
			expLps: []types.LiquidityProviderAccount{
				{
					Address:  addr1,
					Mintable: mintable1,
				},
				{
					Address:  addr2,
					Mintable: mintable2,
				},
			},
		},
		"empty list": {
			setupFunc: func() {
				keeper.RevokeLiquidityProviderAccount(ctx,accAddr1)
				keeper.RevokeLiquidityProviderAccount(ctx, accAddr2)
			},
			expLps: []types.LiquidityProviderAccount(nil),
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			if spec.setupFunc != nil {
				spec.setupFunc()
			}
			gotRsp, gotErr := queryClient.List(sdk.WrapSDKContext(ctx), req)
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			require.NotNil(t, gotRsp)
			sort.Slice(spec.expLps, func(i, j int) bool {
				return spec.expLps[i].Address < spec.expLps[j].Address
			})
			assert.Equal(t, spec.expLps, gotRsp.LiquidityProviders)
		})
	}
}

func TestQueryMintable(t *testing.T) {
	encConfig := MakeTestEncodingConfig()
	initialSupply := sdk.NewCoins(
		sdk.NewCoin("eeur", sdk.NewIntWithDecimal(50, 2)),
		sdk.NewCoin("ejpy", sdk.NewInt(250)),
	)
	ctx, _, _, keeper := createTestComponents(t, initialSupply)

	accAddr1 := sdk.AccAddress(rand.Bytes(sdk.AddrLen))
	addr := accAddr1.String()

	defaultMintable := sdk.NewCoins(
		sdk.NewCoin("eeur", sdk.NewIntWithDecimal(1000, 2)),
	)

	_, err := keeper.CreateLiquidityProvider(ctx, accAddr1, defaultMintable)
	assert.NoError(t, err)

	defaultLP := keeper.GetLiquidityProviderAccount(ctx, accAddr1)
	assert.NotNil(t, defaultLP)

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, encConfig.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, keeper)
	queryClient := types.NewQueryClient(queryHelper)

	specs := map[string]struct {
		req         *types.QueryMintableRequest
		expMintable sdk.Coins
		expErr      bool
	}{
		"all good": {
			req:   &types.QueryMintableRequest{
				Address: addr,
			},
			expMintable: defaultMintable,
		},
		"empty address": {
			req:   &types.QueryMintableRequest{
			},
			expErr: true,
		},
		"non existent provider": {
			req:   &types.QueryMintableRequest{
				Address: sdk.AccAddress(rand.Bytes(sdk.AddrLen)).String(),
			},
			expMintable: sdk.Coins(nil),
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			gotRsp, gotErr := queryClient.Mintable(sdk.WrapSDKContext(ctx), spec.req)
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			require.NotNil(t, gotRsp)
			assert.Equal(t, spec.expMintable, gotRsp.Mintable)
		})
	}
}