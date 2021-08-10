package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	"github.com/e-money/stargate/networks/emibctokens/x/swap/types"
)

func (k msgServer) SendSend(goCtx context.Context, msg *types.MsgSendSend) (*types.MsgSendSendResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: logic before transmitting the packet

	// Construct the packet
	var packet types.SendPacketData

	packet.AmountDenom = msg.AmountDenom
	packet.Recipient = msg.Recipient

	// Transmit the packet
	err := k.TransmitSendPacket(
		ctx,
		packet,
		msg.Port,
		msg.ChannelID,
		clienttypes.ZeroHeight(),
		msg.TimeoutTimestamp,
	)
	if err != nil {
		return nil, err
	}

	return &types.MsgSendSendResponse{}, nil
}
