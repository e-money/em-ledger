package types

// IBC events
const (
	EventTypeTimeout = "timeout"
	// this line is used by starport scaffolding # ibc/packet/event
	EventTypeSendPacket = "send_packet"

	AttributeKeyAckSuccess = "success"
	AttributeKeyAck        = "acknowledgement"
	AttributeKeyAckError   = "error"
)
