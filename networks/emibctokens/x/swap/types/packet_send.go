package types

// ValidateBasic is used for validating the packet
func (p SendPacketData) ValidateBasic() error {

	// TODO: Validate the packet data

	return nil
}

// GetBytes is a helper for serialising
func (p SendPacketData) GetBytes() ([]byte, error) {
	var modulePacket SwapPacketData

	modulePacket.Packet = &SwapPacketData_SendPacket{&p}

	return modulePacket.Marshal()
}
