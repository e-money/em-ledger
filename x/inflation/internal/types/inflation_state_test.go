package types

import (
	"fmt"
	"testing"
)

func TestNewParams1(t *testing.T) {
	config := NewParams("", "caps", "0.04", "kredits", "0.10")
	fmt.Println(config)
}
