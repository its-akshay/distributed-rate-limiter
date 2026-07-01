package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newTestContext creates a gin.Context backed by an httptest.ResponseRecorder.
func newTestContext(method, path string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, nil)
	return c, w
}

func TestNewHealthHandler(t *testing.T) {
	type args struct {
		db    *pgxpool.Pool
		redis *redis.Client
	}
	tests := []struct {
		name string
		args args
		want *HealthHandler
	}{
		{
			name: "creates handler with nil dependencies",
			args: args{db: nil, redis: nil},
			want: &HealthHandler{db: nil, redis: nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewHealthHandler(tt.args.db, tt.args.redis); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewHealthHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHealthHandler_Health(t *testing.T) {
	type fields struct {
		db    *pgxpool.Pool
		redis *redis.Client
	}
	tests := []struct {
		name           string
		fields         fields
		wantStatusCode int
		wantBody       map[string]string
	}{
		{
			name:           "returns 200 with status ok and current version",
			fields:         fields{db: nil, redis: nil},
			wantStatusCode: http.StatusOK,
			wantBody:       map[string]string{"status": "ok", "version": "v5"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := newTestContext(http.MethodGet, "/health")
			h := &HealthHandler{
				db:    tt.fields.db,
				redis: tt.fields.redis,
			}
			h.Health(c)

			if w.Code != tt.wantStatusCode {
				t.Errorf("Health() status code = %v, want %v", w.Code, tt.wantStatusCode)
			}
			var body map[string]string
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("Health() failed to parse response body: %v", err)
			}
			if !reflect.DeepEqual(body, tt.wantBody) {
				t.Errorf("Health() body = %v, want %v", body, tt.wantBody)
			}
		})
	}
}

func TestHealthHandler_Ready(t *testing.T) {
	// Create a pool pointing at a port guaranteed to refuse connections so that
	// Ping returns an error without touching a real database.
	badPool, err := pgxpool.New(context.Background(), "postgres://test:test@127.0.0.1:1/testdb?connect_timeout=1")
	if err != nil {
		t.Skipf("could not construct bad pool for testing: %v", err)
	}
	defer badPool.Close()

	type fields struct {
		db    *pgxpool.Pool
		redis *redis.Client
	}
	tests := []struct {
		name           string
		fields         fields
		wantStatusCode int
		wantBody       map[string]string
	}{
		{
			name:           "returns 503 when database is unreachable",
			fields:         fields{db: badPool, redis: nil},
			wantStatusCode: http.StatusServiceUnavailable,
			wantBody:       map[string]string{"status": "postgres unavailable"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := newTestContext(http.MethodGet, "/ready")
			h := &HealthHandler{
				db:    tt.fields.db,
				redis: tt.fields.redis,
			}
			h.Ready(c)

			if w.Code != tt.wantStatusCode {
				t.Errorf("Ready() status code = %v, want %v", w.Code, tt.wantStatusCode)
			}
			var body map[string]string
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("Ready() failed to parse response body: %v", err)
			}
			if !reflect.DeepEqual(body, tt.wantBody) {
				t.Errorf("Ready() body = %v, want %v", body, tt.wantBody)
			}
		})
	}
}
