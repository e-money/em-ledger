package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
	"github.com/e-money/stargate/networks/emibctokens/x/swap/types"
)

// VoucherDenom returns the voucher of the denom from the port ID and channel ID
// and saves the origin in the store
func VoucherDenom(port string, channel string, denom string) string {
	// since SendPacket did not prefix the denomination, we must prefix denomination here
	sourcePrefix := ibctransfertypes.GetDenomPrefix(port, channel)

	// NOTE: sourcePrefix contains the trailing "/"
	prefixedDenom := sourcePrefix + denom

	// construct the denomination trace from the full raw denomination
	denomTrace := ibctransfertypes.ParseDenomTrace(prefixedDenom)

	voucher := denomTrace.IBCDenom()

	return voucher[:16]
}

// SaveVoucherDenom saves the voucher denom to be able to convert it back later
func (k Keeper) SaveVoucherDenom(ctx sdk.Context, port string, channel string, denom string) {
	voucher := VoucherDenom(port, channel, denom)

	// Store the origin denom
	_, saved := k.GetDenomTrace(ctx, voucher)
	if !saved {
		k.SetDenomTrace(ctx, types.DenomTrace{
			Index:   voucher,
			Port:    port,
			Channel: channel,
			Origin:  denom,
		})
	}
}

// OriginalDenom returns back the original denom of the voucher
// False is returned if the port ID and channel ID provided are not the origins of the voucher
func (k Keeper) OriginalDenom(ctx sdk.Context, port string, channel string, voucher string) (string, bool) {
	trace, exist := k.GetDenomTrace(ctx, voucher)
	if exist {
		// Check if original port and channel
		if trace.Port == port && trace.Channel == channel {
			return trace.Origin, true
		}
	}

	// Not the original chain
	return "", false
}
