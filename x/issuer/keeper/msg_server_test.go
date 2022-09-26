package keeper

import (
	"context"
	"errors"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/issuer/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abcitypes "github.com/tendermint/tendermint/abci/types"
)

var accAddress = sdk.AccAddress("emoney1n5ggspeff4fxc87dvmg0ematr3qzw5l4v20mdv")

func TestIncreaseMintableAmountOfLiquidityProvider(t *testing.T) {
	var (
		issuerAddr               = accAddress
		lpAddr                   = accAddress
		gotIssuer                sdk.AccAddress
		gotLiquidityProviderAddr string
		gotMintableIncrease      sdk.Coins
	)

	keeper := issuerKeeperMock{}
	svr := NewMsgServerImpl(&keeper)

	specs := map[string]struct {
		setup     func(ctx sdk.Context)
		req       *types.MsgIncreaseMintable
		mockFn    func(ctx sdk.Context, liquidityProvider, issuer sdk.AccAddress, mintableIncrease sdk.Coins) (*sdk.Result, error)
		expErr    bool
		expEvents sdk.Events
	}{
		"all good": {
			req: &types.MsgIncreaseMintable{
				Issuer:            issuerAddr.String(),
				LiquidityProvider: lpAddr.String(),
				MintableIncrease:  sdk.NewCoins(sdk.Coin{Denom: "eeur", Amount: sdk.OneInt()}),
			},
			mockFn: func(ctx sdk.Context, liquidityProvider, issuer sdk.AccAddress, mintableIncrease sdk.Coins) (*sdk.Result, error) {
				gotLiquidityProviderAddr = liquidityProvider.String()
				gotIssuer = issuer
				gotMintableIncrease = mintableIncrease
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
		"issuer missing": {
			req: &types.MsgIncreaseMintable{
				LiquidityProvider: lpAddr.String(),
				MintableIncrease:  sdk.NewCoins(sdk.Coin{Denom: "eeur", Amount: sdk.OneInt()}),
			},
			expErr: true,
		},
		"issuer invalid": {
			req: &types.MsgIncreaseMintable{
				Issuer:            "invalid",
				LiquidityProvider: lpAddr.String(),
				MintableIncrease:  sdk.NewCoins(sdk.Coin{Denom: "eeur", Amount: sdk.OneInt()}),
			},
			expErr: true,
		},
		"liquidity provider missing": {
			req: &types.MsgIncreaseMintable{
				Issuer:           issuerAddr.String(),
				MintableIncrease: sdk.NewCoins(sdk.Coin{Denom: "eeur", Amount: sdk.OneInt()}),
			},
			expErr: true,
		},
		"liquidity provider invalid": {
			req: &types.MsgIncreaseMintable{
				Issuer:            issuerAddr.String(),
				LiquidityProvider: "invalid",
				MintableIncrease:  sdk.NewCoins(sdk.Coin{Denom: "eeur", Amount: sdk.OneInt()}),
			},
			expErr: true,
		},
		"processing failure": {
			req: &types.MsgIncreaseMintable{
				Issuer:            issuerAddr.String(),
				LiquidityProvider: lpAddr.String(),
				MintableIncrease:  sdk.NewCoins(sdk.Coin{Denom: "eeur", Amount: sdk.OneInt()}),
			},
			mockFn: func(ctx sdk.Context, liquidityProvider, issuer sdk.AccAddress, mintableIncrease sdk.Coins) (*sdk.Result, error) {
				return nil, errors.New("testing")
			},
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			keeper.IncreaseMintableAmountOfLiquidityProviderFn = spec.mockFn
			eventManager := sdk.NewEventManager()
			ctx := sdk.Context{}.WithContext(context.Background()).WithEventManager(eventManager)
			_, gotErr := svr.IncreaseMintable(sdk.WrapSDKContext(ctx), spec.req)
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, spec.expEvents, eventManager.Events())
			assert.Equal(t, spec.req.LiquidityProvider, gotLiquidityProviderAddr)
			assert.Equal(t, spec.req.Issuer, gotIssuer.String())
			assert.Equal(t, spec.req.MintableIncrease, gotMintableIncrease)
		})
	}
}

func TestDecreaseMintableAmountOfLiquidityProvider(t *testing.T) {
	var (
		issuerAddr               = accAddress
		lpAddr                   = accAddress
		gotIssuer                sdk.AccAddress
		gotLiquidityProviderAddr string
		gotMintableDecrease      sdk.Coins
	)

	keeper := issuerKeeperMock{}
	svr := NewMsgServerImpl(&keeper)

	specs := map[string]struct {
		setup     func(ctx sdk.Context)
		req       *types.MsgDecreaseMintable
		mockFn    func(ctx sdk.Context, liquidityProvider, issuer sdk.AccAddress, mintableDecrease sdk.Coins) (*sdk.Result, error)
		expErr    bool
		expEvents sdk.Events
	}{
		"all good": {
			req: &types.MsgDecreaseMintable{
				Issuer:            issuerAddr.String(),
				LiquidityProvider: lpAddr.String(),
				MintableDecrease:  sdk.NewCoins(sdk.Coin{Denom: "eeur", Amount: sdk.OneInt()}),
			},
			mockFn: func(ctx sdk.Context, liquidityProvider, issuer sdk.AccAddress, mintableDecrease sdk.Coins) (*sdk.Result, error) {
				gotLiquidityProviderAddr = liquidityProvider.String()
				gotIssuer = issuer
				gotMintableDecrease = mintableDecrease
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
		"issuer missing": {
			req: &types.MsgDecreaseMintable{
				LiquidityProvider: lpAddr.String(),
				MintableDecrease:  sdk.NewCoins(sdk.Coin{Denom: "eeur", Amount: sdk.OneInt()}),
			},
			expErr: true,
		},
		"issuer invalid": {
			req: &types.MsgDecreaseMintable{
				Issuer:            "invalid",
				LiquidityProvider: lpAddr.String(),
				MintableDecrease:  sdk.NewCoins(sdk.Coin{Denom: "eeur", Amount: sdk.OneInt()}),
			},
			expErr: true,
		},
		"liquidity provider missing": {
			req: &types.MsgDecreaseMintable{
				Issuer:           issuerAddr.String(),
				MintableDecrease: sdk.NewCoins(sdk.Coin{Denom: "eeur", Amount: sdk.OneInt()}),
			},
			expErr: true,
		},
		"liquidity provider invalid": {
			req: &types.MsgDecreaseMintable{
				Issuer:            issuerAddr.String(),
				LiquidityProvider: "invalid",
				MintableDecrease:  sdk.NewCoins(sdk.Coin{Denom: "eeur", Amount: sdk.OneInt()}),
			},
			expErr: true,
		},
		"processing failure": {
			req: &types.MsgDecreaseMintable{
				Issuer:            issuerAddr.String(),
				LiquidityProvider: lpAddr.String(),
				MintableDecrease:  sdk.NewCoins(sdk.Coin{Denom: "eeur", Amount: sdk.OneInt()}),
			},
			mockFn: func(ctx sdk.Context, liquidityProvider, issuer sdk.AccAddress, mintableDecrease sdk.Coins) (*sdk.Result, error) {
				return nil, errors.New("testing")
			},
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			keeper.DecreaseMintableAmountOfLiquidityProviderFn = spec.mockFn
			eventManager := sdk.NewEventManager()
			ctx := sdk.Context{}.WithContext(context.Background()).WithEventManager(eventManager)
			_, gotErr := svr.DecreaseMintable(sdk.WrapSDKContext(ctx), spec.req)
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, spec.expEvents, eventManager.Events())
			assert.Equal(t, spec.req.LiquidityProvider, gotLiquidityProviderAddr)
			assert.Equal(t, spec.req.Issuer, gotIssuer.String())
			assert.Equal(t, spec.req.MintableDecrease, gotMintableDecrease)
		})
	}
}

func TestSetInflationRate(t *testing.T) {
	var (
		issuerAddr       = accAddress
		gotIssuer        sdk.AccAddress
		gotInflationRate sdk.Dec
		gotDenom         string
	)

	keeper := issuerKeeperMock{}
	svr := NewMsgServerImpl(&keeper)

	captureArgsMock := func(ctx sdk.Context, issuer sdk.AccAddress, inflationRate sdk.Dec, denom string) (*sdk.Result, error) {
		gotIssuer, gotInflationRate, gotDenom = issuer, inflationRate, denom
		return &sdk.Result{}, nil
	}
	specs := map[string]struct {
		setup     func(ctx sdk.Context)
		req       *types.MsgSetInflation
		mockFn    func(ctx sdk.Context, issuer sdk.AccAddress, inflationRate sdk.Dec, denom string) (*sdk.Result, error)
		expErr    bool
		expEvents sdk.Events
	}{
		"all good": {
			req: &types.MsgSetInflation{
				Issuer:        issuerAddr.String(),
				Denom:         "alx",
				InflationRate: sdk.OneDec(),
			},
			mockFn: func(ctx sdk.Context, issuer sdk.AccAddress, inflationRate sdk.Dec, denom string) (*sdk.Result, error) {
				captureArgsMock(ctx, issuer, inflationRate, denom)
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
		"issuer missing": {
			req: &types.MsgSetInflation{
				Denom:         "alx",
				InflationRate: sdk.OneDec(),
			},
			expErr: true,
		},
		"issuer invalid": {
			req: &types.MsgSetInflation{
				Issuer:        "invalid",
				Denom:         "alx",
				InflationRate: sdk.OneDec(),
			},
			expErr: true,
		},
		"denom missing": {
			req: &types.MsgSetInflation{
				Issuer:        issuerAddr.String(),
				InflationRate: sdk.OneDec(),
			},
			mockFn: captureArgsMock,
			expErr: false,
		},
		"denom invalid": {
			req: &types.MsgSetInflation{
				Issuer:        issuerAddr.String(),
				Denom:         "!@#$",
				InflationRate: sdk.OneDec(),
			},
			mockFn: captureArgsMock,
			expErr: false,
		},
		"inflation rate missing": {
			req: &types.MsgSetInflation{
				Issuer: issuerAddr.String(),
				Denom:  "alx",
			},
			mockFn: captureArgsMock,
			expErr: false,
		},
		"inflation rate invalid": {
			req: &types.MsgSetInflation{
				Issuer:        issuerAddr.String(),
				Denom:         "alx",
				InflationRate: sdk.NewDec(-12312),
			},
			mockFn: captureArgsMock,
			expErr: false,
		},
		"processing failure": {
			req: &types.MsgSetInflation{
				Issuer:        issuerAddr.String(),
				Denom:         "alx",
				InflationRate: sdk.OneDec(),
			},
			mockFn: func(ctx sdk.Context, issuer sdk.AccAddress, inflationRate sdk.Dec, denom string) (*sdk.Result, error) {
				return nil, errors.New("testing")
			},
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			keeper.SetInflationRateFn = spec.mockFn
			eventManager := sdk.NewEventManager()
			ctx := sdk.Context{}.WithContext(context.Background()).WithEventManager(eventManager)
			_, gotErr := svr.SetInflation(sdk.WrapSDKContext(ctx), spec.req)
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, spec.req.Issuer, gotIssuer.String())
			assert.Equal(t, spec.req.Denom, gotDenom)
			assert.Equal(t, spec.req.InflationRate, gotInflationRate)

			if len(spec.expEvents) == 0 && len(eventManager.Events()) == 0 {
				return // not fail when nil != empty
			}
			assert.Equal(t, spec.expEvents, eventManager.Events())
		})
	}
}

type issuerKeeperMock struct {
	IncreaseMintableAmountOfLiquidityProviderFn func(ctx sdk.Context, liquidityProvider, issuer sdk.AccAddress, mintableIncrease sdk.Coins) (*sdk.Result, error)
	DecreaseMintableAmountOfLiquidityProviderFn func(ctx sdk.Context, liquidityProvider, issuer sdk.AccAddress, mintableDecrease sdk.Coins) (*sdk.Result, error)
	RevokeLiquidityProviderFn                   func(ctx sdk.Context, liquidityProvider, issuerAddress sdk.AccAddress) (*sdk.Result, error)
	SetInflationRateFn                          func(ctx sdk.Context, issuer sdk.AccAddress, inflationRate sdk.Dec, denom string) (*sdk.Result, error)
}

func (m issuerKeeperMock) IncreaseMintableAmountOfLiquidityProvider(ctx sdk.Context, liquidityProvider, issuer sdk.AccAddress, mintableIncrease sdk.Coins) (*sdk.Result, error) {
	if m.IncreaseMintableAmountOfLiquidityProviderFn == nil {
		panic("not expected to be called")
	}
	return m.IncreaseMintableAmountOfLiquidityProviderFn(ctx, liquidityProvider, issuer, mintableIncrease)
}

func (m issuerKeeperMock) DecreaseMintableAmountOfLiquidityProvider(ctx sdk.Context, liquidityProvider, issuer sdk.AccAddress, mintableDecrease sdk.Coins) (*sdk.Result, error) {
	if m.DecreaseMintableAmountOfLiquidityProviderFn == nil {
		panic("not expected to be called")
	}
	return m.DecreaseMintableAmountOfLiquidityProviderFn(ctx, liquidityProvider, issuer, mintableDecrease)
}

func (m issuerKeeperMock) RevokeLiquidityProvider(ctx sdk.Context, liquidityProvider, issuerAddress sdk.AccAddress) (*sdk.Result, error) {
	if m.RevokeLiquidityProviderFn == nil {
		panic("not expected to be called")
	}
	return m.RevokeLiquidityProviderFn(ctx, liquidityProvider, issuerAddress)
}

func (m issuerKeeperMock) SetInflationRate(ctx sdk.Context, issuer sdk.AccAddress, inflationRate sdk.Dec, denom string) (*sdk.Result, error) {
	if m.SetInflationRateFn == nil {
		panic("not expected to be called")
	}
	return m.SetInflationRateFn(ctx, issuer, inflationRate, denom)
}
