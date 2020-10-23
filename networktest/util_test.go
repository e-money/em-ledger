// This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
//
// Please contact partners@e-money.com for licensing related questions.

// +build bdd

package networktest

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestScanner1(t *testing.T) {
	waiter, scanner := createOutputScanner("55", 2*time.Second)
	go generateOutputForScanner(scanner)

	start := time.Now()
	require.False(t, waiter()) // Should timeout.
	diff := time.Now().Sub(start).Seconds()
	require.Equal(t, 2, int(diff)) // Do some rough time verification.
}

func TestScanner2(t *testing.T) {
	waiter, scanner := createOutputScanner("33", time.Hour)
	go generateOutputForScanner(scanner)

	require.True(t, waiter()) // Should not timeout.
}

func TestScannerMultipleTriggers(t *testing.T) {
	_, scanner := createOutputScanner("10", time.Hour)

	scanner("TEN 10")
	scanner("TEN 10")
	scanner("TEN 10")
}

func generateOutputForScanner(scanner func(string)) {
	for i := 0; i < 100; i++ {
		scanner(fmt.Sprintf("This is line number %02d", i))
		time.Sleep(200 * time.Millisecond)
	}
}
