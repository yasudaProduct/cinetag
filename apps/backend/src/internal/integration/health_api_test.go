//go:build integration

package integration

import (
	"testing"

	"cinetag-backend/src/internal/testutil"
)

// GET /health
// ステータス 200 と {"status":"ok"} を返すことを確認する。
func TestHealthEndpoint(t *testing.T) {
	env := setupTestEnv(t)

	resp := env.request("GET", "/health", nil, nil)
	resp.AssertStatus(t, 200)

	body := resp.JSON(t)
	testutil.AssertJSON(t, body, map[string]any{
		"status": "ok",
	})
}
