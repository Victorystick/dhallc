package add

import "testing"

//go:generate go run ../../cmd/dhallc.go --out ./add.go ./add.dhall

func TestSum(t *testing.T) {
	if Sum != 3 {
		t.Errorf("Expected 3, got %d\n", Sum)
	}
}
