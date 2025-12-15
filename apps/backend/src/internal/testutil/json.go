package testutil

import (
	"encoding/json"
	"testing"
)

// MustMarshalJSON は v を JSON に変換し、失敗したらテストを落とします。
func MustMarshalJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal json: %v", err)
	}
	return b
}

// MustUnmarshalJSON は JSON を dst に詰め、失敗したらテストを落とします。
func MustUnmarshalJSON(t *testing.T, b []byte, dst any) {
	t.Helper()
	if err := json.Unmarshal(b, dst); err != nil {
		t.Fatalf("failed to unmarshal json: %v", err)
	}
}
