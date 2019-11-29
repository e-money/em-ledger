// This software is Copyright (c) 2019 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

package types

// Slashing module event types
var (
	EventTypeSlash         = "slash"
	EventTypeLiveness      = "liveness"
	EventTypePenaltyPayout = "penalty_payout"

	AttributeKeyAddress      = "address"
	AttributeKeyHeight       = "height"
	AttributeKeyPower        = "power"
	AttributeKeyReason       = "reason"
	AttributeKeyJailed       = "jailed"
	AttributeKeyMissedBlocks = "missed_blocks"
	AttributeKeyAmount       = "amount"

	AttributeValueDoubleSign       = "double_sign"
	AttributeValueMissingSignature = "missing_signature"
	AttributeValueCategory         = ModuleName
)
