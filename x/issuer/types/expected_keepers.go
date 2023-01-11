package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type (
	InflationKeeper interface {
		SetInflation(sdk.Context, sdk.Dec, string) (*sdk.Result, error)
		AddDenoms(sdk.Context, []string) (*sdk.Result, error)
	}

	BankKeeper interface {
		GetDenomMetaData(ctx sdk.Context, denom string) banktypes.Metadata
		SetDenomMetaData(ctx sdk.Context, denomMetaData banktypes.Metadata)
	}
)
