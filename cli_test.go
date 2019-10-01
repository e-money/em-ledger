// +build bdd

package emoney

import (
	"context"
	"fmt"
	"testing"
	"time"

	"emoney/network_test"
)

func TestTestnet(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	testnet := network_test.NewTestnetWithContext(ctx)

	err := testnet.Setup()
	if err != nil {
		fmt.Println(err)
		t.Fail()
		return
	}

	err = testnet.Start()
	if err != nil {
		fmt.Println(err)
		t.Fail()
		return
	}

	time.Sleep(30 * time.Second)

	testnet.Restart()

	time.Sleep(30 * time.Second)

	cancel()

	time.Sleep(10 * time.Second)
}
