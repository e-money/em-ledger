package keeper

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/authority/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abcitypes "github.com/tendermint/tendermint/abci/types"
)

func TestCreateIssuer(t *testing.T) {
	var (
		authorityAddr = mustParseAddress("emoney1kt0vh0ttget0xx77g6d3ttnvq2lnxx6vp3uyl0")
		issuerAddr    = mustParseAddress("emoney17up20gamd0vh6g9ne0uh67hx8xhyfrv2lyazgu")
		gotAuthority  sdk.AccAddress
		gotIssuer     sdk.AccAddress
		gotDenoms     []types.Denomination
	)

	keeper := authorityKeeperMock{}
	svr := NewMsgServerImpl(&keeper)

	specs := map[string]struct {
		setup     func(ctx sdk.Context)
		req       *types.MsgCreateIssuer
		mockFn    func(ctx sdk.Context, authority sdk.AccAddress, issuerAddress sdk.AccAddress, denoms []types.Denomination) (*sdk.Result, error)
		expErr    bool
		expEvents sdk.Events
	}{
		"all good": {
			req: &types.MsgCreateIssuer{
				Authority:     authorityAddr.String(),
				Issuer:        issuerAddr.String(),
				Denominations: []types.Denomination{{Base: "foo"}, {Base: "bar"}},
			},
			mockFn: func(ctx sdk.Context, authority sdk.AccAddress, issuerAddress sdk.AccAddress, denoms []types.Denomination) (*sdk.Result, error) {
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
				Denominations: []types.Denomination{{Base: "foo"}, {Base: "bar"}},
			},
			expErr: true,
		},
		"authority invalid": {
			req: &types.MsgCreateIssuer{
				Authority:     "invalid",
				Issuer:        issuerAddr.String(),
				Denominations: []types.Denomination{{Base: "foo"}, {Base: "bar"}},
			},
			expErr: true,
		},
		"issuer missing": {
			req: &types.MsgCreateIssuer{
				Authority:     authorityAddr.String(),
				Denominations: []types.Denomination{{Base: "foo"}, {Base: "bar"}},
			},
			expErr: true,
		},
		"issuer invalid": {
			req: &types.MsgCreateIssuer{
				Authority:     authorityAddr.String(),
				Issuer:        "invalid",
				Denominations: []types.Denomination{{Base: "foo"}, {Base: "bar"}},
			},
			expErr: true,
		},
		"processing failure": {
			req: &types.MsgCreateIssuer{
				Authority:     authorityAddr.String(),
				Issuer:        issuerAddr.String(),
				Denominations: []types.Denomination{{Base: "foo"}, {Base: "bar"}},
			},
			mockFn: func(ctx sdk.Context, authority sdk.AccAddress, issuerAddress sdk.AccAddress, denoms []types.Denomination) (*sdk.Result, error) {
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

func TestScheduleUpgrade(t *testing.T) {
	var (
		authorityAddr = mustParseAddress("emoney1kt0vh0ttget0xx77g6d3ttnvq2lnxx6vp3uyl0")
		gotAuthority  sdk.AccAddress
		gotPlan       upgradetypes.Plan
	)

	keeper := authorityKeeperMock{}
	svr := NewMsgServerImpl(&keeper)

	specs := map[string]struct {
		req       *types.MsgScheduleUpgrade
		mockFn    func(ctx sdk.Context, authority sdk.AccAddress, plan upgradetypes.Plan) (*sdk.Result, error)
		expErr    bool
		expEvents sdk.Events
	}{
		"all good": {
			req: &types.MsgScheduleUpgrade{
				Authority: authorityAddr.String(),
				Plan: upgradetypes.Plan{
					Name:   "plan8",
					Height: 100,
				},
			},
			mockFn: func(ctx sdk.Context, authority sdk.AccAddress, plan upgradetypes.Plan) (*sdk.Result, error) {
				gotAuthority, gotPlan = authority, plan
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
			req: &types.MsgScheduleUpgrade{
				Plan: upgradetypes.Plan{
					Name:   "test1",
					Height: 100,
				},
			},
			expErr: true,
		},
		"authority invalid": {
			req: &types.MsgScheduleUpgrade{
				Authority: "invalid",
				Plan: upgradetypes.Plan{
					Name:   "test1",
					Height: 100,
				},
			},
			expErr: true,
		},
		"invalid height value": {
			req: &types.MsgScheduleUpgrade{
				Authority: authorityAddr.String(),
				Plan: upgradetypes.Plan{
					Name:   "test1",
					Height: -100,
				},
			},
			mockFn: func(ctx sdk.Context, authority sdk.AccAddress, plan upgradetypes.Plan) (*sdk.Result, error) {
				return nil, errors.New("testing")
			},
			expErr: true,
		},
		"missing height": {
			req: &types.MsgScheduleUpgrade{
				Authority: authorityAddr.String(),
				Plan: upgradetypes.Plan{
					Name: "test1",
				},
			},
			mockFn: func(ctx sdk.Context, authority sdk.AccAddress, plan upgradetypes.Plan) (*sdk.Result, error) {
				return nil, errors.New("testing")
			},
			expErr: true,
		},
		"missing plan name": {
			req: &types.MsgScheduleUpgrade{
				Authority: authorityAddr.String(),
				Plan: upgradetypes.Plan{
					Height: 1,
				},
			},
			mockFn: func(ctx sdk.Context, authority sdk.AccAddress, plan upgradetypes.Plan) (*sdk.Result, error) {
				return nil, errors.New("testing")
			},
			expErr: true,
		},
		"featuring both time and height": {
			req: &types.MsgScheduleUpgrade{
				Authority: authorityAddr.String(),
				Plan: upgradetypes.Plan{
					Name:   "test1",
					Time:   time.Now(),
					Height: 1,
				},
			},
			mockFn: func(ctx sdk.Context, authority sdk.AccAddress, plan upgradetypes.Plan) (*sdk.Result, error) {
				return nil, errors.New("testing")
			},
			expErr: true,
		},
	}

	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			keeper.scheduleUpgradefn = spec.mockFn
			eventManager := sdk.NewEventManager()
			ctx := sdk.Context{}.WithContext(context.Background()).WithEventManager(eventManager)
			_, gotErr := svr.ScheduleUpgrade(sdk.WrapSDKContext(ctx), spec.req)
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, spec.expEvents, eventManager.Events())
			assert.Equal(t, spec.req.Authority, gotAuthority.String())
			assert.Equal(t, spec.req.Plan, gotPlan)
		})
	}
}

func TestReplaceAuth(t *testing.T) {
	var (
		authorityAddr                 = mustParseAddress("emoney1kt0vh0ttget0xx77g6d3ttnvq2lnxx6vp3uyl0")
		newAuthorityAddr              = mustParseAddress("emoney1hq6tnhqg4t7358f3vd9crru93lv0cgekdxrtgv")
		gotAuthority, gotNewAuthority sdk.AccAddress
	)

	keeper := authorityKeeperMock{}
	svr := NewMsgServerImpl(&keeper)

	specs := map[string]struct {
		req       *types.MsgReplaceAuthority
		mockFn    func(ctx sdk.Context, authority, newAuthority sdk.AccAddress) (*sdk.Result, error)
		expErr    bool
		expEvents sdk.Events
	}{
		"all good": {
			req: &types.MsgReplaceAuthority{
				Authority:    authorityAddr.String(),
				NewAuthority: newAuthorityAddr.String(),
			},
			mockFn: func(ctx sdk.Context, authority, newAuthority sdk.AccAddress) (*sdk.Result, error) {
				gotAuthority, gotNewAuthority = authority, newAuthority
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
		"Same authority (fallback -> current)": {
			req: &types.MsgReplaceAuthority{
				Authority:    authorityAddr.String(),
				NewAuthority: authorityAddr.String(),
			},
			mockFn: func(ctx sdk.Context, authority, newAuthority sdk.AccAddress) (*sdk.Result, error) {
				gotAuthority, gotNewAuthority = authority, newAuthority
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
			req: &types.MsgReplaceAuthority{
				NewAuthority: newAuthorityAddr.String(),
			},
			expErr: true,
		},
		"new authority missing": {
			req: &types.MsgReplaceAuthority{
				Authority: newAuthorityAddr.String(),
			},
			expErr: true,
		},
		"authority invalid": {
			req: &types.MsgReplaceAuthority{
				Authority:    "invalid",
				NewAuthority: newAuthorityAddr.String(),
			},
			expErr: true,
		},
		"new authority invalid": {
			req: &types.MsgReplaceAuthority{
				Authority:    authorityAddr.String(),
				NewAuthority: "invalid",
			},
			expErr: true,
		},
		"processing failure": {
			req: &types.MsgReplaceAuthority{
				Authority:    authorityAddr.String(),
				NewAuthority: newAuthorityAddr.String(),
			},
			mockFn: func(ctx sdk.Context, authority, newAuthority sdk.AccAddress) (*sdk.Result, error) {
				return nil, errors.New("testing")
			},
			expErr: true,
		},
	}

	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			keeper.replaceAuthorityfn = spec.mockFn
			eventManager := sdk.NewEventManager()
			ctx := sdk.Context{}.WithContext(context.Background()).WithEventManager(eventManager)
			_, gotErr := svr.ReplaceAuthority(sdk.WrapSDKContext(ctx), spec.req)
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, spec.expEvents, eventManager.Events())
			assert.Equal(t, spec.req.Authority, gotAuthority.String())
			assert.Equal(t, spec.req.NewAuthority, gotNewAuthority.String())
		})
	}
}

func TestGrpcSetParams(t *testing.T) {
	var (
		authorityAddr = mustParseAddress("emoney1kt0vh0ttget0xx77g6d3ttnvq2lnxx6vp3uyl0")
		gotAuthority  sdk.AccAddress
		gotChanges    []proposal.ParamChange
	)

	keeper := authorityKeeperMock{}
	svr := NewMsgServerImpl(&keeper)

	specs := map[string]struct {
		req       *types.MsgSetParameters
		mockFn    func(ctx sdk.Context, authority sdk.AccAddress, changes []proposal.ParamChange) (*sdk.Result, error)
		expErr    bool
		expEvents sdk.Events
	}{
		"all good": {
			req: &types.MsgSetParameters{
				Authority: authorityAddr.String(),
				Changes: []proposal.ParamChange{
					{
						Subspace: "staking",
						Key:      "MaxValidators",
						Value:    "10",
					}},
			},
			mockFn: func(ctx sdk.Context, authority sdk.AccAddress, changes []proposal.ParamChange) (*sdk.Result, error) {
				gotAuthority = authority
				gotChanges = changes

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
		"processing failure": {
			req: &types.MsgSetParameters{
				Authority: authorityAddr.String(),
				Changes: []proposal.ParamChange{
					{
						Subspace: "staking",
						Key:      "MaxValidators",
						Value:    "Ten",
					}},
			},
			mockFn: func(ctx sdk.Context, authority sdk.AccAddress, changes []proposal.ParamChange) (*sdk.Result, error) {
				return nil, errors.New("testing")
			},
			expErr: true,
		},
	}

	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			keeper.setParamsfn = spec.mockFn
			eventManager := sdk.NewEventManager()
			ctx := sdk.Context{}.WithContext(context.Background()).WithEventManager(eventManager)
			_, gotErr := svr.SetParameters(sdk.WrapSDKContext(ctx), spec.req)
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, spec.expEvents, eventManager.Events())
			assert.Equal(t, spec.req.Authority, gotAuthority.String())
			assert.Equal(t, spec.req.Changes, gotChanges)
		})
	}
}

// mock implementation of authorityKeeper interface
type authorityKeeperMock struct {
	createIssuerfn     func(ctx sdk.Context, authority sdk.AccAddress, issuerAddress sdk.AccAddress, denoms []types.Denomination) (*sdk.Result, error)
	destroyIssuerfn    func(ctx sdk.Context, authority sdk.AccAddress, issuerAddress sdk.AccAddress) (*sdk.Result, error)
	SetGasPricesfn     func(ctx sdk.Context, authority sdk.AccAddress, gasprices sdk.DecCoins) (*sdk.Result, error)
	replaceAuthorityfn func(ctx sdk.Context, authority, newAuthority sdk.AccAddress) (*sdk.Result, error)
	scheduleUpgradefn  func(ctx sdk.Context, authority sdk.AccAddress, plan upgradetypes.Plan) (*sdk.Result, error)
	getUpgradePlanfn   func(ctx sdk.Context) (plan upgradetypes.Plan, havePlan bool)
	applyUpgradefn     func(ctx sdk.Context, authority sdk.AccAddress, plan upgradetypes.Plan) (*sdk.Result, error)
	setParamsfn        func(ctx sdk.Context, authority sdk.AccAddress, changes []proposal.ParamChange) (*sdk.Result, error)
}

func (a authorityKeeperMock) SetParams(
	ctx sdk.Context, authority sdk.AccAddress, changes []proposal.ParamChange,
) (*sdk.Result, error) {
	if a.setParamsfn == nil {
		panic("not expected to be called")
	}

	return a.setParamsfn(ctx, authority, changes)
}

func (a authorityKeeperMock) createIssuer(ctx sdk.Context, authority sdk.AccAddress, issuerAddress sdk.AccAddress, denoms []types.Denomination) (*sdk.Result, error) {
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

func (a authorityKeeperMock) ScheduleUpgrade(
	ctx sdk.Context, authority sdk.AccAddress, plan upgradetypes.Plan,
) (*sdk.Result, error) {
	if a.scheduleUpgradefn == nil {
		panic("not expected to be called")
	}

	return a.scheduleUpgradefn(ctx, authority, plan)
}

func (a authorityKeeperMock) GetUpgradePlan(ctx sdk.Context) (plan upgradetypes.Plan, havePlan bool) {
	if a.getUpgradePlanfn == nil {
		panic("not expected to be called")
	}

	return a.getUpgradePlanfn(ctx)
}

func (a authorityKeeperMock) ApplyUpgrade(
	ctx sdk.Context, authority sdk.AccAddress, plan upgradetypes.Plan,
) (*sdk.Result, error) {
	if a.applyUpgradefn == nil {
		panic("not expected to be called")
	}

	return a.applyUpgradefn(ctx, authority, plan)
}
