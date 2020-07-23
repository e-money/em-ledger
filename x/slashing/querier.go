// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package slashing

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/e-money/em-ledger/x/slashing/types"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewQuerier creates a new querier for slashing clients.
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case QueryParameters:
			return queryParams(ctx, k)
		case QuerySigningInfo:
			return querySigningInfo(ctx, req, k)
		case QuerySigningInfos:
			return querySigningInfos(ctx, req, k)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown %s query endpoint: %s", types.ModuleName, path[0])
		}
	}
}

func queryParams(ctx sdk.Context, k Keeper) ([]byte, error) {
	params := k.GetParams(ctx)

	res, err := codec.MarshalJSONIndent(ModuleCdc, params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func querySigningInfo(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params QuerySigningInfoParams

	err := ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	signingInfo, found := k.getValidatorSigningInfo(ctx, params.ConsAddress)
	if !found {
		return nil, sdkerrors.Wrap(types.ErrNoSigningInfoFound, params.ConsAddress.String())
	}

	res, err := codec.MarshalJSONIndent(ModuleCdc, signingInfo)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func querySigningInfos(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params QuerySigningInfosParams

	err := ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	var signingInfos []ValidatorSigningInfo

	k.IterateValidatorSigningInfos(func(consAddr sdk.ConsAddress, info ValidatorSigningInfo) (stop bool) {
		signingInfos = append(signingInfos, info)
		return false
	})

	start, end := client.Paginate(len(signingInfos), params.Page, params.Limit, int(k.sk.MaxValidators(ctx)))
	if start < 0 || end < 0 {
		signingInfos = []ValidatorSigningInfo{}
	} else {
		signingInfos = signingInfos[start:end]
	}

	res, err := codec.MarshalJSONIndent(ModuleCdc, signingInfos)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}
