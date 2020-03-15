package password

import (
	"fmt"
	"testing"
)

func TestRandPassword(t *testing.T) {
	fmt.Println(len(RandPassword()))
}
