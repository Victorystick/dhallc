package list

import "testing"

//go:generate go run ../../cmd/dhallc.go --out ./list.go ./list.dhall

func TestNats(t *testing.T) {
	expected := []uint{1, 2, 3, 4, 3, 0}

	if len(Nats) != len(expected) {
		t.Errorf("Expected %d elements got, got %v\n", len(expected), Nats)
	}

	for i, v := range expected {
		if Nats[i] != v {
			t.Errorf("Item at index %d didn't match %v != %v\n", i, Nats[i], v)
		}
	}

	if NatLen != 4 {
		t.Errorf("Expected 4 elements got, got %v\n", Nats)
	}
}

func TestInts(t *testing.T) {
	if len(Ints) != 4 {
		t.Errorf("Expected 4 elements got, got %v\n", Ints)
	}
}
