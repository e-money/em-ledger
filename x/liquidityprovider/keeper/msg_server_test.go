package keeper

import (
	"context"
	"errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/liquidityprovider/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/rand"
	"testing"
)

func TestMintTokens(t *testing.T) {
	var (
		lpAddr                   = randomAddress()
		gotLiquidityProviderAddr string
		gotAmount                sdk.Coins
	)

	keeper := lpKeeperMock{}
	svr := NewMsgServerImpl(&keeper)

	specs := map[string]struct {
		req       *types.MsgMintTokens
		mockFn    func(ctx sdk.Context, liquidityProvider sdk.AccAddress, amount sdk.Coins) (*sdk.Result, error)
		expErr    bool
		expEvents sdk.Events
	}{
		"all good": {
			req: &types.MsgMintTokens{
				LiquidityProvider: lpAddr,
				Amount:            sdk.NewCoins(sdk.Coin{Denom: "eeur", Amount: sdk.OneInt()}),
			},
			mockFn: func(ctx sdk.Context, liquidityProvider sdk.AccAddress, amount sdk.Coins) (*sdk.Result, error) {
				gotLiquidityProviderAddr = liquidityProvider.String()
				gotAmount = amount
				return &sdk.Result{
					Events: []abcitypes.Event{{
						Type:       "testing",
						Attributes: []abcitypes.EventAttribute{{Key: []byte("foo"), Value: []byte("bar")}},
					}},
				}, nil
			},
			expEvents: sdk.Events{{
				Type:       "testing",
				Attributes: []abcitypes.EventAttribute{{Key: []byte("foo"), Value: []byte("bar")}},
			}},
		},
		"liquidity provider missing": {
			req: &types.MsgMintTokens{
				Amount: sdk.NewCoins(sdk.Coin{Denom: "eeur", Amount: sdk.OneInt()}),
			},
			expErr: true,
		},
		"liquidity provider invalid": {
			req: &types.MsgMintTokens{
				LiquidityProvider: "invalid",
				Amount:            sdk.NewCoins(sdk.Coin{Denom: "eeur", Amount: sdk.OneInt()}),
			},
			expErr: true,
		},
		"Amount missing": {
			mockFn: func(ctx sdk.Context, liquidityProvider sdk.AccAddress, amount sdk.Coins) (*sdk.Result, error) {
				gotLiquidityProviderAddr, gotAmount = liquidityProvider.String(), amount
				return &sdk.Result{}, nil
			},
			req: &types.MsgMintTokens{
				LiquidityProvider: lpAddr,
			},
			expErr:    false,
			expEvents: []sdk.Event{},
		},
		"processing failure": {
			req: &types.MsgMintTokens{
				LiquidityProvider: lpAddr,
				Amount:            sdk.NewCoins(sdk.Coin{Denom: "eeur", Amount: sdk.OneInt()}),
			},
			mockFn: func(ctx sdk.Context, liquidityProvider sdk.AccAddress, amount sdk.Coins) (*sdk.Result, error) {
				return nil, errors.New("testing")
			},
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			keeper.MintTokensFn = spec.mockFn
			eventManager := sdk.NewEventManager()
			ctx := sdk.Context{}.WithContext(context.Background()).WithEventManager(eventManager)
			_, gotErr := svr.MintTokens(sdk.WrapSDKContext(ctx), spec.req)
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, spec.expEvents, eventManager.Events())
			assert.Equal(t, spec.req.LiquidityProvider, gotLiquidityProviderAddr)
			assert.Equal(t, spec.req.GetAmount(), gotAmount)
		})
	}
}

func TestBurnTokens(t *testing.T) {
	var (
		lpAddr                   = randomAddress()
		gotLiquidityProviderAddr string
		gotAmount                sdk.Coins
	)

	keeper := lpKeeperMock{}
	svr := NewMsgServerImpl(&keeper)

	specs := map[string]struct {
		req       *types.MsgBurnTokens
		mockFn    func(ctx sdk.Context, liquidityProvider sdk.AccAddress, amount sdk.Coins) (*sdk.Result, error)
		expErr    bool
		expEvents sdk.Events
	}{
		"all good": {
			req: &types.MsgBurnTokens{
				LiquidityProvider: lpAddr,
				Amount:            sdk.NewCoins(sdk.Coin{Denom: "eeur", Amount: sdk.OneInt()}),
			},
			mockFn: func(ctx sdk.Context, liquidityProvider sdk.AccAddress, amount sdk.Coins) (*sdk.Result, error) {
				gotLiquidityProviderAddr = liquidityProvider.String()
				gotAmount = amount
				return &sdk.Result{
					Events: []abcitypes.Event{{
						Type:       "testing",
						Attributes: []abcitypes.EventAttribute{{Key: []byte("foo"), Value: []byte("bar")}},
					}},
				}, nil
			},
			expEvents: sdk.Events{{
				Type:       "testing",
				Attributes: []abcitypes.EventAttribute{{Key: []byte("foo"), Value: []byte("bar")}},
			}},
		},
		"liquidity provider missing": {
			req: &types.MsgBurnTokens{
				Amount: sdk.NewCoins(sdk.Coin{Denom: "eeur", Amount: sdk.OneInt()}),
			},
			expErr: true,
		},
		"liquidity provider invalid": {
			req: &types.MsgBurnTokens{
				LiquidityProvider: "invalid",
				Amount:            sdk.NewCoins(sdk.Coin{Denom: "eeur", Amount: sdk.OneInt()}),
			},
			expErr: true,
		},
		"Amount missing": {
			mockFn: func(ctx sdk.Context, liquidityProvider sdk.AccAddress, amount sdk.Coins) (*sdk.Result, error) {
				gotLiquidityProviderAddr, gotAmount = liquidityProvider.String(), amount
				return &sdk.Result{}, nil
			},
			req: &types.MsgBurnTokens{
				LiquidityProvider: lpAddr,
			},
			expErr:    false,
			expEvents: []sdk.Event{},
		},
		"processing failure": {
			req: &types.MsgBurnTokens{
				LiquidityProvider: lpAddr,
				Amount:            sdk.NewCoins(sdk.Coin{Denom: "eeur", Amount: sdk.OneInt()}),
			},
			mockFn: func(ctx sdk.Context, liquidityProvider sdk.AccAddress, amount sdk.Coins) (*sdk.Result, error) {
				return nil, errors.New("testing")
			},
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			keeper.BurnTokensFromBalanceFn = spec.mockFn
			eventManager := sdk.NewEventManager()
			ctx := sdk.Context{}.WithContext(context.Background()).WithEventManager(eventManager)
			_, gotErr := svr.BurnTokens(sdk.WrapSDKContext(ctx), spec.req)
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, spec.expEvents, eventManager.Events())
			assert.Equal(t, spec.req.LiquidityProvider, gotLiquidityProviderAddr)
			assert.Equal(t, spec.req.GetAmount(), gotAmount)
		})
	}
}

type lpKeeperMock struct {
	MintTokensFn            func(ctx sdk.Context, liquidityProvider sdk.AccAddress, amount sdk.Coins) (*sdk.Result, error)
	BurnTokensFromBalanceFn func(ctx sdk.Context, liquidityProvider sdk.AccAddress, amount sdk.Coins) (*sdk.Result, error)
}

func (m lpKeeperMock) MintTokens(ctx sdk.Context, liquidityProvider sdk.AccAddress, amount sdk.Coins) (*sdk.Result, error) {
	if m.MintTokensFn == nil {
		panic("not expected to be called")
	}
	return m.MintTokensFn(ctx, liquidityProvider, amount)
}

func (m lpKeeperMock) BurnTokensFromBalance(ctx sdk.Context, liquidityProvider sdk.AccAddress, amount sdk.Coins) (*sdk.Result, error) {
	if m.BurnTokensFromBalanceFn == nil {
		panic("not expected to be called")
	}
	return m.BurnTokensFromBalanceFn(ctx, liquidityProvider, amount)
}

func randomAddress() string {
	return sdk.AccAddress(rand.Bytes(legacyAddrLen)).String()
}
