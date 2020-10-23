// +build fast_consensus

package keeper

import (
	"time"
)

func init() {
	// Update on every block
	updateInterval = time.Millisecond
}
