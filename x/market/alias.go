// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package market

import (
	"github.com/e-money/em-ledger/x/market/client/cli"
	"github.com/e-money/em-ledger/x/market/keeper"
	"github.com/e-money/em-ledger/x/market/types"
)

const (
	ModuleName       = types.ModuleName
	RouterKey        = types.RouterKey
	StoreKey         = types.StoreKey
	StoreKeyIdx      = types.StoreKeyIdx
	QuerierRoute     = types.QuerierRoute
	QueryByAccount   = types.QueryByAccount
	QueryInstrument  = types.QueryInstrument
	QueryInstruments = types.QueryInstruments

	Order_Limit  = types.Order_Limit
	Order_Market = types.Order_Market

	TimeInForce_GoodTillCancel    = types.TimeInForce_GoodTilCancel
	TimeInForce_ImmediateOrCancel = types.TimeInForce_ImmediateOrCancel
	TimeInForce_FillOrKill        = types.TimeInForce_FillOrKill
)

var (
	ModuleCdc    = types.ModuleCdc
	NewKeeper    = keeper.NewKeeper
	BeginBlocker = keeper.BeginBlocker
	NewOrder     = types.NewOrder

	ErrClientOrderIdNotFound                   = types.ErrClientOrderIdNotFound
	ErrOrderInstrumentChanged                  = types.ErrOrderInstrumentChanged
	ErrNoSourceRemaining                       = types.ErrNoSourceRemaining
	ErrUnknownAsset                            = types.ErrUnknownAsset
	ErrAccountBalanceInsufficient              = types.ErrAccountBalanceInsufficient
	ErrInvalidInstrument                       = types.ErrInvalidInstrument
	ErrInvalidPrice                            = types.ErrInvalidPrice
	ErrAccountBalanceInsufficientForInstrument = types.ErrAccountBalanceInsufficientForInstrument
	ErrNonUniqueClientOrderId                  = types.ErrNonUniqueClientOrderId

	EventTypeCancel = types.EventTypeCancel

	EmitNewOrderEvent = types.EmitNewOrderEvent
	EmitFilledEvent   = types.EmitFilledEvent
	EmitCancelEvent   = types.EmitCancelEvent

	GetOwnerKey               = types.GetOwnerKey
	GetMarketDataKey          = types.GetMarketDataKey
	GetMarketDataPrefix       = types.GetMarketDataPrefix
	GetPriorityKey            = types.GetPriorityKey
	GetPriorityKeyBySrcAndDst = types.GetPriorityKeyBySrcAndDst
	GetOrderIDGeneratorKey    = types.GetOrderIDGeneratorKey

	GetTxCmd    = cli.GetTxCmd
	GetQueryCmd = cli.GetQueryCmd
)

type (
	Keeper        = keeper.Keeper
	Order         = types.Order
	MarketData    = types.MarketData
	ExecutionPlan = types.ExecutionPlan

	QueryInstrumentResponse         = keeper.QueryInstrumentResponse
	QueryInstrumentsResponse        = keeper.QueryInstrumentsResponse
	QueryInstrumentsWrapperResponse = keeper.QueryInstrumentsWrapperResponse

	MsgAddMarketOrder          = types.MsgAddMarketOrder
	MsgAddLimitOrder           = types.MsgAddLimitOrder
	MsgCancelOrder             = types.MsgCancelOrder
	MsgCancelReplaceLimitOrder = types.MsgCancelReplaceLimitOrder

	AccountKeeper    = types.AccountKeeper
	BankKeeper       = types.BankKeeper
	SupplyKeeper     = types.SupplyKeeper
	RestrictedKeeper = types.RestrictedKeeper
)
