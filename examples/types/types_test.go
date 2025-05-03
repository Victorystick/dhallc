package types

import "testing"

//go:generate go run ../../cmd/dhallc.go --out ./types.go ./types.dhall

func TestConfigs(t *testing.T) {
	expected := []string{"bill", "jane"}

	if len(Configs) != len(expected) {
		t.Errorf("Expected %d elements got, got %v\n", len(expected), Configs)
	}

	for i, v := range expected {
		if Configs[i].publicKey != "/home/"+v+"/.ssh/id_ed25519.pub" {
			t.Errorf("Item at index %d didn't match %v != %v\n", i, Configs[i], v)
		}
	}
}
