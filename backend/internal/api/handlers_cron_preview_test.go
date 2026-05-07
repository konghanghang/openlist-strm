package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func newCronPreviewRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	server := &Server{}
	router := gin.New()
	router.POST("/cron/preview", server.handleCronPreview)
	return router
}

func postCronPreview(t *testing.T, router *gin.Engine, body interface{}) *httptest.ResponseRecorder {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/cron/preview", &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestHandleCronPreview_Valid6Field(t *testing.T) {
	router := newCronPreviewRouter()

	w := postCronPreview(t, router, CronPreviewRequest{
		CronExpr: "0 */5 * * * *",
		Count:    3,
	})

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp CronPreviewResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.NextRunTimes) != 3 {
		t.Fatalf("len(NextRunTimes) = %d, want 3", len(resp.NextRunTimes))
	}

	now := time.Now()
	for i, ts := range resp.NextRunTimes {
		if !ts.After(now) {
			t.Errorf("NextRunTimes[%d] = %v, expected to be after %v", i, ts, now)
		}
	}
	for i := 1; i < len(resp.NextRunTimes); i++ {
		if !resp.NextRunTimes[i].After(resp.NextRunTimes[i-1]) {
			t.Errorf("NextRunTimes not strictly increasing: %v then %v",
				resp.NextRunTimes[i-1], resp.NextRunTimes[i])
		}
	}
}

func TestHandleCronPreview_DefaultCount(t *testing.T) {
	router := newCronPreviewRouter()

	w := postCronPreview(t, router, CronPreviewRequest{
		CronExpr: "0 0 2 * * *",
	})

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", w.Code, w.Body.String())
	}

	var resp CronPreviewResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.NextRunTimes) != 3 {
		t.Errorf("default count should be 3, got %d", len(resp.NextRunTimes))
	}
}

func TestHandleCronPreview_Reject5Field(t *testing.T) {
	router := newCronPreviewRouter()

	// 5 字段格式（无 seconds），与 createMapping 严格对齐——必须拒绝
	w := postCronPreview(t, router, CronPreviewRequest{
		CronExpr: "*/5 * * * *",
		Count:    3,
	})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for 5-field expression, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestHandleCronPreview_InvalidExpression(t *testing.T) {
	router := newCronPreviewRouter()

	w := postCronPreview(t, router, CronPreviewRequest{
		CronExpr: "not a cron",
	})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleCronPreview_EmptyExpression(t *testing.T) {
	router := newCronPreviewRouter()

	w := postCronPreview(t, router, CronPreviewRequest{
		CronExpr: "",
	})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty expr, got %d", w.Code)
	}
}

func TestHandleCronPreview_CountBounds(t *testing.T) {
	router := newCronPreviewRouter()

	cases := []struct {
		name    string
		count   int
		wantOK  bool
		wantLen int
	}{
		{"count=1 ok", 1, true, 1},
		{"count=10 ok", 10, true, 10},
		{"count=11 reject", 11, false, 0},
		{"count=-1 reject", -1, false, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := postCronPreview(t, router, CronPreviewRequest{
				CronExpr: "0 0 * * * *",
				Count:    tc.count,
			})
			if tc.wantOK {
				if w.Code != http.StatusOK {
					t.Fatalf("status = %d, body=%s", w.Code, w.Body.String())
				}
				var resp CronPreviewResponse
				_ = json.Unmarshal(w.Body.Bytes(), &resp)
				if len(resp.NextRunTimes) != tc.wantLen {
					t.Errorf("len = %d, want %d", len(resp.NextRunTimes), tc.wantLen)
				}
			} else {
				if w.Code != http.StatusBadRequest {
					t.Errorf("expected 400, got %d", w.Code)
				}
			}
		})
	}
}
