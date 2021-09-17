package encoder

import "testing"

func TestEncoder(t *testing.T) {
	var id uint64 = 8912323

	encoded := Encode(id)
	t.Logf("Encoded value: %s", encoded)

	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Error decoding: %v", err)
	}
	if decoded != id {
		t.Fatalf("Decode value mismatch. expected: %d, got: %d", id, decoded)
	}
	t.Logf("Decoded back: %d", decoded)
}
