package types

import "time"

const (
	ModuleName   = "authority"
	StoreKey     = ModuleName
	RouterKey    = ModuleName
	QuerierRoute = ModuleName

	// Query endpoints supported by the authority querier
	QueryGasPrices = "gasprices"

	// AuthorityTransitionDuration is the period during which the former
	// authority and new authority are in effect. During this period the former
	// acts like a backup authority and cannot change till expiration.
	AuthorityTransitionDuration = 24 * time.Hour
)
