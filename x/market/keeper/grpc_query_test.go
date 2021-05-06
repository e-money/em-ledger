package keeper

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/e-money/em-ledger/x/market/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestQueryByAccount(t *testing.T) {
	enc := MakeTestEncodingConfig()
	ctx, k, _, _ := createTestComponentsWithEncoding(t, enc)

	var myAddress = randomAccAddress()

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, enc.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, k)
	queryClient := types.NewQueryClient(queryHelper)
	o, err := types.NewOrder(
		ctx.BlockTime(),
		types.TimeInForce_GoodTillCancel,
		sdk.NewCoin("alx", sdk.OneInt()),
		sdk.NewCoin("blx", sdk.OneInt()),
		myAddress, "myOrderID",
	)
	require.NoError(t, err)
	k.setOrder(ctx, &o)

	expectedPlusOne := o
	expectedPlusOne.Created.Add(1*time.Second)

	specs := map[string]struct {
		req      *types.QueryByAccountRequest
		expErr   bool
		expState []*types.Order
		// Ensure date is set correctly
		createdPlusOne bool
	}{
		"all good": {
			req:      &types.QueryByAccountRequest{Address: myAddress.String()},
			expState: []*types.Order{&o},
		},
		"created plus a sec": {
			req:      &types.QueryByAccountRequest{Address: myAddress.String()},
			expState: []*types.Order{&expectedPlusOne},
		},
		"empty address": {
			req:    &types.QueryByAccountRequest{Address: ""},
			expErr: true,
		},
		"invalid address": {
			req:    &types.QueryByAccountRequest{Address: "invalid"},
			expErr: true,
		},
		"unknown address": {
			req: &types.QueryByAccountRequest{Address: randomAccAddress().String()},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			gotRsp, gotErr := queryClient.ByAccount(sdk.WrapSDKContext(ctx), spec.req)
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)

			if spec.createdPlusOne {
				assert.NotEqual(t, spec.expState, gotRsp.Orders)
				// set equal
				gotRsp.Orders[0].Created.Add(1*time.Second)
			}

			assert.Equal(t, spec.expState, gotRsp.Orders)
		})
	}
}

func TestInstruments(t *testing.T) {
	enc := MakeTestEncodingConfig()
	ctx, k, _, bk := createTestComponentsWithEncoding(t, enc)

	bk.SetSupply(ctx, banktypes.NewSupply(coins("1alx,1blx")))

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, enc.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, k)
	queryClient := types.NewQueryClient(queryHelper)

	specs := map[string]struct {
		req      *types.QueryInstrumentsRequest
		expState []types.QueryInstrumentsResponse_Element
	}{
		"all good": {
			req: &types.QueryInstrumentsRequest{},
			expState: []types.QueryInstrumentsResponse_Element{
				{Source: "alx", Destination: "blx"},
				{Source: "blx", Destination: "alx"},
			},
		},
		"nil param": {
			expState: []types.QueryInstrumentsResponse_Element{
				{Source: "alx", Destination: "blx"},
				{Source: "blx", Destination: "alx"},
			},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			gotRsp, gotErr := queryClient.Instruments(sdk.WrapSDKContext(ctx), spec.req)
			require.NoError(t, gotErr)
			require.NotNil(t, gotRsp)
			assert.Equal(t, spec.expState, gotRsp.Instruments)
		})
	}
}
func TestInstrument(t *testing.T) {
	enc := MakeTestEncodingConfig()
	ctx, k, ak, bk := createTestComponentsWithEncoding(t, enc)

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, enc.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, k)
	queryClient := types.NewQueryClient(queryHelper)

	acc := createAccount(ctx, ak, bk, randomAddress(), "1000usd")

	o := order(ctx.BlockTime(), acc, "100usd", "100chf")
	_, err := k.NewOrderSingle(ctx, o)
	require.NoError(t, err)

	oPlusOne := order(
		ctx.BlockTime().Add(time.Second), acc, "100usd", "100gbp",
	)
	_, err = k.NewOrderSingle(ctx, oPlusOne)
	require.NoError(t, err)

	specs := map[string]struct {
		req      *types.QueryInstrumentRequest
		expErr   bool
		expState *types.QueryInstrumentResponse
	}{
		"all good": {
			req: &types.QueryInstrumentRequest{Source: "usd", Destination: "chf"},
			expState: &types.QueryInstrumentResponse{
				Source:      "usd",
				Destination: "chf",
				Orders: []types.QueryOrderResponse{
					{
						Owner:           acc.GetAddress().String(),
						SourceRemaining: "100",
						Price:           sdk.NewDec(1),
						Created:         ctx.BlockTime(),
					},
				},
			},
		},
		"created plus a sec": {
			req: &types.QueryInstrumentRequest{Source: "usd", Destination: "gbp"},
			expState: &types.QueryInstrumentResponse{
				Source:      "usd",
				Destination: "gbp",
				Orders: []types.QueryOrderResponse{
					{
						ID:              1,
						Owner:           acc.GetAddress().String(),
						SourceRemaining: "100",
						Price:           sdk.NewDec(1),
						Created:         ctx.BlockTime().Add(1 * time.Second),
					},
				},
			},
		},
		"invalid demon": {
			req:    &types.QueryInstrumentRequest{Source: "#!@@", Destination: "chf"},
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			gotRsp, gotErr := queryClient.Instrument(sdk.WrapSDKContext(ctx), spec.req)
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, spec.expState, gotRsp)
		})
	}
}
