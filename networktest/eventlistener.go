// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

// +build bdd

package networktest

import (
	"context"
	"fmt"
	"sync"
	"time"

	bep3types "github.com/e-money/bep3/module/types"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	client "github.com/tendermint/tendermint/rpc/client/http"
	ct "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"
)

type EventListener struct {
	client *client.HTTP
}

func NewEventListener() (EventListener, error) {
	return NewEventListenerNode(0)
}

func NewEventListenerNode(node int) (EventListener, error) {
	rpcPorts :=[]int{26657, 26660, 26662, 26664}

	url := fmt.Sprintf("http://localhost:%d", rpcPorts[node])
	httpClient, err := client.New(url, "/websocket")
	if err != nil {
		return EventListener{}, err
	}

	if err := httpClient.Start(); err != nil {
		return EventListener{}, err
	}

	return EventListener{client: httpClient}, nil
}

func (el EventListener) subscribeQuery(query string, listener func(ct.ResultEvent) bool) error {
	return el.subscribeQueryDuration(query, time.Minute, listener)
}

func (el EventListener) subscribeQueryDuration(query string, timeout time.Duration, listener func(ct.ResultEvent) bool) error {
	ctx := context.Background()
	eventChannel, err := el.client.WSEvents.Subscribe(ctx, "", query, 1000)
	if err != nil {
		return err
	}

	defer el.client.WSEvents.Unsubscribe(ctx, "", query)

	// process found events till timeout or listener false value
	for {
		select {
		case evt := <-eventChannel:
			if !listener(evt) {
				return nil
			}
		case <-time.After(timeout):
			err := fmt.Errorf("timeout waiting for event:%s", query)
			fmt.Println("***", err.Error())
			return err
		}
	}
}

func (el EventListener) awaitQuery(query string) (func() *ct.ResultEvent, error) {
	return el.awaitQueryDuration(query, time.Minute)
}

func (el EventListener) awaitQueryDuration(query string, timeout time.Duration) (func() *ct.ResultEvent, error) {
	ctx := context.Background()
	eventChannel, err := el.client.WSEvents.Subscribe(ctx, "", query, 1000)
	if err != nil {
		return nil, err
	}

	res := func() *ct.ResultEvent {
		defer el.client.WSEvents.Unsubscribe(ctx, "", query)

		select {
		case evt := <-eventChannel:
			return &evt
		case <-time.After(timeout):
			fmt.Println(" *** Timeout waiting for event.", query)
			return nil
		}
	}

	return res, nil
}

func (el EventListener) AwaitNewBlock() (func() bool, error) {
	eventFn, err := el.awaitQuery("tm.event='NewBlock'")
	if err != nil {
		return nil, err
	}

	return func() bool {
		event := eventFn()

		if event == nil {
			return false
		}

		_, ok := event.Data.(types.EventDataNewBlock)
		if !ok {
			fmt.Printf("Unexpected event type data received: %T\n", event.Data)
			return false
		}

		return true
	}, nil
}

// SubscribeExpirations fetches all new block events till a timeout or the
// expected Bep3 expiration occurs.
func (el EventListener) SubscribeExpirations(swapID string, timeout time.Duration) error {
	return el.subscribeQueryDuration(
		"tm.event='NewBlock'", timeout,
		func(event ct.ResultEvent) bool {
			eventNB, ok := event.Data.(types.EventDataNewBlock)
			if !ok {
				// fetch next
				return true
			}
			for _, e := range eventNB.ResultBeginBlock.Events {
				if e.Type == bep3types.EventTypeSwapsExpired {
					key := string(e.Attributes[0].Key)
					if key == bep3types.AttributeKeyAtomicSwapIDs {
						// attribute value is a list; omit brackets
						val := string(e.Attributes[0].Value[1 : len(e.Attributes[0].Value)-1])
						if val == swapID {
							// do not fetch additional events.
							return false
						}
					}
				}
			}

			// fetch next
			return true
		},
	)
}

// SubTx fetches all tx events till a timeout or the expected transactions'
// hashes (txHashes) are found withing the emitted events. This is thread safe
// and if called from a single thread you may pass a nil mutex.
func (el EventListener) SubTx(
	mu *sync.RWMutex, txHashes map[string]bool, total int32, timeout time.Duration,
) (found int32, err error) {
	ctx := context.Background()
	query := "tm.event='Tx'"
	eventChannel, err := el.client.WSEvents.Subscribe(ctx, "", query, 10000)
	if err != nil {
		return 0, err
	}

	defer el.client.WSEvents.Unsubscribe(ctx, "", query)

	for ;found < total;{
		select {
		case evt := <-eventChannel:
			for k, v := range evt.Events {
				if k == "tx.hash" {
					for _, hash := range v {
						// writes may still be occurring
						mu.RLock()
						in := txHashes[hash]
						mu.RUnlock()
						if in {
							found++
						}
					}
				}
			}
		case <-time.After(timeout):
			err := fmt.Errorf("timeout waiting for event:%s", query)
			fmt.Println("***", err.Error())
			return found, err
		}
	}

	return found, nil
}

func (el EventListener) AwaitPenaltyPayout() (func() bool, error) {
	eventFn, err := el.awaitQuery("tm.event='NewBlock' AND penalty_payout.address CONTAINS 'emoney'")
	if err != nil {
		return nil, err
	}

	return func() bool {
		evt := eventFn()
		return evt != nil
	}, nil
}

func (el EventListener) AwaitSlash() (func() *abcitypes.Event, error) {
	// todo (reviewer): please note the new query format. See https://docs.cosmos.network/master/core/events.html#subscribing-to-events
	eventFn, err := el.awaitQuery("tm.event='NewBlock' AND slash.reason='missing_signature'")
	if err != nil {
		return nil, err
	}

	return func() *abcitypes.Event {
		event := eventFn()

		if event == nil {
			return nil
		}

		newBlockEventData, ok := event.Data.(types.EventDataNewBlock)
		if !ok {
			fmt.Printf("Unexpected event type data received: %T\n", event.Data)
			return nil
		}

		for _, e := range newBlockEventData.ResultBeginBlock.Events {
			if e.Type == "slash" {
				return &e
			}
		}

		return nil
	}, nil
}

func (el EventListener) AwaitValidatorSetChange() (func() *types.EventDataValidatorSetUpdates, error) {
	eventFn, err := el.awaitQuery("tm.event='ValidatorSetUpdates'")
	if err != nil {
		return nil, err
	}

	return func() *types.EventDataValidatorSetUpdates {
		event := eventFn()

		validatorSetUpdates, ok := event.Data.(types.EventDataValidatorSetUpdates)
		if ok {
			return &validatorSetUpdates
		}

		return nil
	}, nil
}
