// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

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
		GetDenomMetaData(ctx sdk.Context, denom string) (banktypes.Metadata, bool)
		SetDenomMetaData(ctx sdk.Context, denomMetaData banktypes.Metadata)
	}
)
