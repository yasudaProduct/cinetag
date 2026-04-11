package testutil

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"
)

// HTTPResponse は httptest.ResponseRecorder をラップし、アサーション用ヘルパーを提供します。
type HTTPResponse struct {
	Recorder *httptest.ResponseRecorder
}

// AssertStatus はレスポンスのステータスコードを検証します。
func (r *HTTPResponse) AssertStatus(t *testing.T, expected int) {
	t.Helper()
	if r.Recorder.Code != expected {
		t.Fatalf("expected status %d, got %d; body=%s", expected, r.Recorder.Code, r.Recorder.Body.String())
	}
}

// JSON はレスポンスボディを JSON としてパースして返します。
func (r *HTTPResponse) JSON(t *testing.T) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal(r.Recorder.Body.Bytes(), &out); err != nil {
		t.Fatalf("レスポンスの JSON パースに失敗: %v; body=%s", err, r.Recorder.Body.String())
	}
	return out
}

// AssertJSON は JSON レスポンスの各フィールドを期待値と比較します。
// expected の各キーについて、値が一致するか検証します。
func AssertJSON(t *testing.T, data map[string]any, expected map[string]any) {
	t.Helper()
	for key, want := range expected {
		got, ok := data[key]
		if !ok {
			t.Errorf("key %q が存在しません; body=%v", key, data)
			continue
		}
		if fmt.Sprintf("%v", got) != fmt.Sprintf("%v", want) {
			t.Errorf("key %q: expected %v (%T), got %v (%T)", key, want, want, got, got)
		}
	}
}

// AssertHasKeys は JSON レスポンスに指定のキーがすべて存在することを検証します。
func AssertHasKeys(t *testing.T, data map[string]any, keys ...string) {
	t.Helper()
	for _, key := range keys {
		if _, ok := data[key]; !ok {
			t.Errorf("key %q が存在しません; body=%v", key, data)
		}
	}
}

// AssertListResponse はページネーション付き一覧レスポンスの共通フィールドを検証します。
func AssertListResponse(t *testing.T, data map[string]any, expectedTotal float64, expectedPage float64, expectedPageSize float64) {
	t.Helper()
	AssertJSON(t, data, map[string]any{
		"total_count": expectedTotal,
		"page":        expectedPage,
		"page_size":   expectedPageSize,
	})
	if _, ok := data["items"]; !ok {
		t.Errorf("key \"items\" が存在しません")
	}
}

// GetItems は一覧レスポンスの items を []map[string]any として返します。
func GetItems(t *testing.T, data map[string]any) []map[string]any {
	t.Helper()
	rawItems, ok := data["items"].([]any)
	if !ok {
		t.Fatalf("items が配列ではありません: %v", data["items"])
	}
	items := make([]map[string]any, len(rawItems))
	for i, raw := range rawItems {
		item, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("items[%d] がオブジェクトではありません: %v", i, raw)
		}
		items[i] = item
	}
	return items
}
