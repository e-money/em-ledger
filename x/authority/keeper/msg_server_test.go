package keeper

import (
	"context"
	"errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/authority/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	"testing"
)

func TestCreateIssuer(t *testing.T) {
	var (
		authorityAddr = mustParseAddress("emoney1kt0vh0ttget0xx77g6d3ttnvq2lnxx6vp3uyl0")
		issuerAddr    = mustParseAddress("emoney17up20gamd0vh6g9ne0uh67hx8xhyfrv2lyazgu")
		gotAuthority  sdk.AccAddress
		gotIssuer     sdk.AccAddress
		gotDenoms     []string
	)

	keeper := authorityKeeperMock{}
	svr := NewMsgServerImpl(&keeper)

	specs := map[string]struct {
		setup     func(ctx sdk.Context)
		req       *types.MsgCreateIssuer
		mockFn    func(ctx sdk.Context, authority sdk.AccAddress, issuerAddress sdk.AccAddress, denoms []string) (*sdk.Result, error)
		expErr    bool
		expEvents sdk.Events
	}{
		"all good": {
			req: &types.MsgCreateIssuer{
				Authority:     authorityAddr.String(),
				Issuer:        issuerAddr.String(),
				Denominations: []string{"foo", "bar"},
			},
			mockFn: func(ctx sdk.Context, authority sdk.AccAddress, issuerAddress sdk.AccAddress, denoms []string) (*sdk.Result, error) {
				gotAuthority, gotIssuer, gotDenoms = authority, issuerAddr, denoms
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
		"authority missing": {
			req: &types.MsgCreateIssuer{
				Issuer:        issuerAddr.String(),
				Denominations: []string{"foo", "bar"},
			},
			expErr: true,
		},
		"authority invalid": {
			req: &types.MsgCreateIssuer{
				Authority:     "invalid",
				Issuer:        issuerAddr.String(),
				Denominations: []string{"foo", "bar"},
			},
			expErr: true,
		},
		"issuer missing": {
			req: &types.MsgCreateIssuer{
				Authority:     authorityAddr.String(),
				Denominations: []string{"foo", "bar"},
			},
			expErr: true,
		},
		"issuer invalid": {
			req: &types.MsgCreateIssuer{
				Authority:     authorityAddr.String(),
				Issuer:        "invalid",
				Denominations: []string{"foo", "bar"},
			},
			expErr: true,
		},
		"processing failure": {
			req: &types.MsgCreateIssuer{
				Authority:     authorityAddr.String(),
				Issuer:        issuerAddr.String(),
				Denominations: []string{"foo", "bar"},
			},
			mockFn: func(ctx sdk.Context, authority sdk.AccAddress, issuerAddress sdk.AccAddress, denoms []string) (*sdk.Result, error) {
				return nil, errors.New("testing")
			},
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			keeper.createIssuerfn = spec.mockFn
			eventManager := sdk.NewEventManager()
			ctx := sdk.Context{}.WithContext(context.Background()).WithEventManager(eventManager)
			_, gotErr := svr.CreateIssuer(sdk.WrapSDKContext(ctx), spec.req)
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, spec.expEvents, eventManager.Events())
			assert.Equal(t, spec.req.Authority, gotAuthority.String())
			assert.Equal(t, spec.req.Issuer, gotIssuer.String())
			assert.Equal(t, spec.req.Denominations, gotDenoms)
		})
	}
}

func TestDestroyIssuer(t *testing.T) {
	var (
		authorityAddr = mustParseAddress("emoney1kt0vh0ttget0xx77g6d3ttnvq2lnxx6vp3uyl0")
		issuerAddr    = mustParseAddress("emoney17up20gamd0vh6g9ne0uh67hx8xhyfrv2lyazgu")
		gotAuthority  sdk.AccAddress
		gotIssuer     sdk.AccAddress
	)

	keeper := authorityKeeperMock{}
	svr := NewMsgServerImpl(&keeper)

	specs := map[string]struct {
		req       *types.MsgDestroyIssuer
		mockFn    func(ctx sdk.Context, authority sdk.AccAddress, issuerAddress sdk.AccAddress) (*sdk.Result, error)
		expErr    bool
		expEvents sdk.Events
	}{
		"all good": {
			req: &types.MsgDestroyIssuer{
				Authority: authorityAddr.String(),
				Issuer:    issuerAddr.String(),
			},
			mockFn: func(ctx sdk.Context, authority sdk.AccAddress, issuerAddress sdk.AccAddress) (*sdk.Result, error) {
				gotAuthority, gotIssuer = authority, issuerAddr
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
		"authority missing": {
			req: &types.MsgDestroyIssuer{
				Issuer: issuerAddr.String(),
			},
			expErr: true,
		},
		"authority invalid": {
			req: &types.MsgDestroyIssuer{
				Authority: "invalid",
				Issuer:    issuerAddr.String(),
			},
			expErr: true,
		},
		"issuer missing": {
			req: &types.MsgDestroyIssuer{
				Authority: authorityAddr.String(),
			},
			expErr: true,
		},
		"issuer invalid": {
			req: &types.MsgDestroyIssuer{
				Authority: authorityAddr.String(),
				Issuer:    "invalid",
			},
			expErr: true,
		},
		"processing failure": {
			req: &types.MsgDestroyIssuer{
				Authority: authorityAddr.String(),
				Issuer:    issuerAddr.String(),
			},
			mockFn: func(ctx sdk.Context, authority sdk.AccAddress, issuerAddress sdk.AccAddress) (*sdk.Result, error) {
				return nil, errors.New("testing")
			},
			expErr: true,
		},
	}

	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			keeper.destroyIssuerfn = spec.mockFn
			eventManager := sdk.NewEventManager()
			ctx := sdk.Context{}.WithContext(context.Background()).WithEventManager(eventManager)
			// when
			_, gotErr := svr.DestroyIssuer(sdk.WrapSDKContext(ctx), spec.req)
			// then
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, spec.expEvents, eventManager.Events())
			assert.Equal(t, spec.req.Authority, gotAuthority.String())
			assert.Equal(t, spec.req.Issuer, gotIssuer.String())

		})
	}
}

func TestSetGasPrices(t *testing.T) {
	var (
		authorityAddr = mustParseAddress("emoney1kt0vh0ttget0xx77g6d3ttnvq2lnxx6vp3uyl0")
		gotAuthority  sdk.AccAddress
		gotGasPrices  sdk.DecCoins
	)

	keeper := authorityKeeperMock{}
	svr := NewMsgServerImpl(&keeper)

	specs := map[string]struct {
		req       *types.MsgSetGasPrices
		mockFn    func(ctx sdk.Context, authority sdk.AccAddress, gasprices sdk.DecCoins) (*sdk.Result, error)
		expErr    bool
		expEvents sdk.Events
	}{
		"all good": {
			req: &types.MsgSetGasPrices{
				Authority: authorityAddr.String(),
				GasPrices: sdk.DecCoins{sdk.NewDecCoin("eeur", sdk.OneInt())},
			},
			mockFn: func(ctx sdk.Context, authority sdk.AccAddress, gasprices sdk.DecCoins) (*sdk.Result, error) {
				gotAuthority, gotGasPrices = authority, gasprices
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
		"authority missing": {
			req: &types.MsgSetGasPrices{
				GasPrices: sdk.DecCoins{sdk.NewDecCoin("eeur", sdk.OneInt())},
			},
			expErr: true,
		},
		"authority invalid": {
			req: &types.MsgSetGasPrices{
				Authority: "invalid",
				GasPrices: sdk.DecCoins{sdk.NewDecCoin("eeur", sdk.OneInt())},
			},
			expErr: true,
		},
		"processing failure": {
			req: &types.MsgSetGasPrices{
				Authority: authorityAddr.String(),
				GasPrices: sdk.DecCoins{sdk.NewDecCoin("eeur", sdk.OneInt())},
			},
			mockFn: func(ctx sdk.Context, authority sdk.AccAddress, gasprices sdk.DecCoins) (*sdk.Result, error) {
				return nil, errors.New("testing")
			},
			expErr: true,
		},
	}

	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			keeper.SetGasPricesfn = spec.mockFn
			eventManager := sdk.NewEventManager()
			ctx := sdk.Context{}.WithContext(context.Background()).WithEventManager(eventManager)
			_, gotErr := svr.SetGasPrices(sdk.WrapSDKContext(ctx), spec.req)
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, spec.expEvents, eventManager.Events())
			assert.Equal(t, spec.req.Authority, gotAuthority.String())
			assert.Equal(t, spec.req.GasPrices, gotGasPrices)
		})
	}
}

// mock implementation of authorityKeeper interface
type authorityKeeperMock struct {
	createIssuerfn  func(ctx sdk.Context, authority sdk.AccAddress, issuerAddress sdk.AccAddress, denoms []string) (*sdk.Result, error)
	destroyIssuerfn func(ctx sdk.Context, authority sdk.AccAddress, issuerAddress sdk.AccAddress) (*sdk.Result, error)
	SetGasPricesfn  func(ctx sdk.Context, authority sdk.AccAddress, gasprices sdk.DecCoins) (*sdk.Result, error)
	replaceAuthorityfn func(ctx sdk.Context, authority, newAuthority sdk.AccAddress) (*sdk.Result, error)
}

func (a authorityKeeperMock) createIssuer(ctx sdk.Context, authority sdk.AccAddress, issuerAddress sdk.AccAddress, denoms []string) (*sdk.Result, error) {
	if a.createIssuerfn == nil {
		panic("not expected to be called")
	}
	return a.createIssuerfn(ctx, authority, issuerAddress, denoms)
}

func (a authorityKeeperMock) destroyIssuer(ctx sdk.Context, authority sdk.AccAddress, issuerAddress sdk.AccAddress) (*sdk.Result, error) {
	if a.destroyIssuerfn == nil {
		panic("not expected to be called")
	}
	return a.destroyIssuerfn(ctx, authority, issuerAddress)
}

func (a authorityKeeperMock) SetGasPrices(ctx sdk.Context, authority sdk.AccAddress, gasprices sdk.DecCoins) (*sdk.Result, error) {
	if a.SetGasPricesfn == nil {
		panic("not expected to be called")
	}
	return a.SetGasPricesfn(ctx, authority, gasprices)
}

func (a authorityKeeperMock) replaceAuthority(ctx sdk.Context, authority, newAuthority sdk.AccAddress) (*sdk.Result, error) {
	if a.replaceAuthorityfn == nil {
		panic("not expected to be called")
	}

	return a.replaceAuthorityfn(ctx, authority, newAuthority)
}
