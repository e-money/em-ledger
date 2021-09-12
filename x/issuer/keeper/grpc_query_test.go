package keeper

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/e-money/em-ledger/x/issuer/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/rand"
)

func TestQueryIssuers(t *testing.T) {
	encConfig := MakeTestEncodingConfig()
	ctx, _, _, keeper, _ := createTestComponentsWithEncodingConfig(t, encConfig)
	myIssuers := []types.Issuer{types.NewIssuer(randomAccAddress(), "foo", "bar")}
	keeper.setIssuers(ctx, myIssuers)

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, encConfig.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, keeper)
	queryClient := types.NewQueryClient(queryHelper)

	specs := map[string]struct {
		req        *types.QueryIssuersRequest
		expIssuers []types.Issuer
	}{
		"all good": {
			req:        &types.QueryIssuersRequest{},
			expIssuers: myIssuers,
		},
		"nil param": {
			expIssuers: myIssuers,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			gotRsp, gotErr := queryClient.Issuers(sdk.WrapSDKContext(ctx), spec.req)
			require.NoError(t, gotErr)
			require.NotNil(t, gotRsp)
			assert.Equal(t, spec.expIssuers, gotRsp.Issuers)
		})
	}
}

func randomAccAddress() sdk.AccAddress {
	return rand.Bytes(sdk.AddrLen)
}
