package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/its-akshay/distributed-rate-limiter/internal/model"
)

type mockRuleRepo struct {
	createErr error
	getRule   *model.Rule
	getErr    error
	listRules []model.Rule
	listErr   error

	createCalled bool
	createdRule  *model.Rule
	getCalled    bool
	gotGetID     int64
	listCalled   bool
}

func (m *mockRuleRepo) GetById(ctx context.Context, id int64) (*model.Rule, error) {
	m.getCalled = true
	m.gotGetID = id
	return m.getRule, m.getErr
}

func (m *mockRuleRepo) Create(ctx context.Context, rule *model.Rule) error {
	m.createCalled = true
	m.createdRule = rule
	return m.createErr
}

func (m *mockRuleRepo) List(ctx context.Context) ([]model.Rule, error) {
	m.listCalled = true
	return m.listRules, m.listErr
}

type mockRateLimiterService struct {
	allowed   bool
	err       error
	called    bool
	gotKey    string
	gotRuleID int64
}

func (m *mockRateLimiterService) Check(ctx context.Context, key string, ruleId int64) (bool, error) {
	m.called = true
	m.gotKey = key
	m.gotRuleID = ruleId
	return m.allowed, m.err
}

func newJSONContext(method, path, body string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	return c, w
}

func TestNewRuleHandler(t *testing.T) {
	repo := &mockRuleRepo{}
	svc := &mockRateLimiterService{}

	h := NewRuleHandler(repo, svc)
	if h == nil {
		t.Fatal("NewRuleHandler() returned nil")
	}
	if h.repo != repo {
		t.Fatal("NewRuleHandler() repo dependency was not set")
	}
	if h.service != svc {
		t.Fatal("NewRuleHandler() service dependency was not set")
	}
}

func TestRuleHandler_CreateRule(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		repo           *mockRuleRepo
		wantStatusCode int
		wantError      bool
		wantCreateCall bool
	}{
		{
			name:           "returns 400 when request body is invalid json",
			body:           "{",
			repo:           &mockRuleRepo{},
			wantStatusCode: http.StatusBadRequest,
			wantError:      true,
			wantCreateCall: false,
		},
		{
			name:           "returns 500 when repository create fails",
			body:           `{"name":"api-rule","limit_count":10,"window_seconds":60}`,
			repo:           &mockRuleRepo{createErr: errors.New("create failed")},
			wantStatusCode: http.StatusInternalServerError,
			wantError:      true,
			wantCreateCall: true,
		},
		{
			name:           "returns 201 when rule is created",
			body:           `{"name":"api-rule","limit_count":10,"window_seconds":60}`,
			repo:           &mockRuleRepo{},
			wantStatusCode: http.StatusCreated,
			wantError:      false,
			wantCreateCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := newJSONContext(http.MethodPost, "/rules", tt.body)
			h := &RuleHandler{
				repo:    tt.repo,
				service: &mockRateLimiterService{},
			}
			h.CreateRule(c)

			if w.Code != tt.wantStatusCode {
				t.Fatalf("CreateRule() status code = %v, want %v", w.Code, tt.wantStatusCode)
			}

			if tt.repo.createCalled != tt.wantCreateCall {
				t.Fatalf("CreateRule() create called = %v, want %v", tt.repo.createCalled, tt.wantCreateCall)
			}

			var body map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("CreateRule() failed to parse response body: %v", err)
			}
			_, hasError := body["error"]
			if hasError != tt.wantError {
				t.Fatalf("CreateRule() error field present = %v, want %v", hasError, tt.wantError)
			}
		})
	}
}

func TestRuleHandler_GetRule(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		pathID         string
		repo           *mockRuleRepo
		wantStatusCode int
		wantGetCall    bool
		wantGetID      int64
	}{
		{
			name:           "returns 400 for non numeric id",
			path:           "/rules/abc",
			pathID:         "abc",
			repo:           &mockRuleRepo{},
			wantStatusCode: http.StatusBadRequest,
			wantGetCall:    false,
		},
		{
			name:           "returns 404 when rule does not exist",
			path:           "/rules/1",
			pathID:         "1",
			repo:           &mockRuleRepo{getErr: errors.New("not found")},
			wantStatusCode: http.StatusNotFound,
			wantGetCall:    true,
			wantGetID:      1,
		},
		{
			name:   "returns 200 with rule",
			path:   "/rules/2",
			pathID: "2",
			repo: &mockRuleRepo{getRule: &model.Rule{
				ID:            2,
				Name:          "burst",
				LimitCount:    20,
				WindowSeconds: 60,
			}},
			wantStatusCode: http.StatusOK,
			wantGetCall:    true,
			wantGetID:      2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := newTestContext(http.MethodGet, tt.path)
			c.Params = gin.Params{{Key: "id", Value: tt.pathID}}

			h := &RuleHandler{
				repo:    tt.repo,
				service: &mockRateLimiterService{},
			}
			h.GetRule(c)

			if w.Code != tt.wantStatusCode {
				t.Fatalf("GetRule() status code = %v, want %v", w.Code, tt.wantStatusCode)
			}
			if tt.repo.getCalled != tt.wantGetCall {
				t.Fatalf("GetRule() repo called = %v, want %v", tt.repo.getCalled, tt.wantGetCall)
			}
			if tt.wantGetCall && tt.repo.gotGetID != tt.wantGetID {
				t.Fatalf("GetRule() repo id = %v, want %v", tt.repo.gotGetID, tt.wantGetID)
			}
		})
	}
}

func TestRuleHandler_ListRules(t *testing.T) {
	tests := []struct {
		name           string
		repo           *mockRuleRepo
		wantStatusCode int
		wantListCall   bool
		wantBody       any
	}{
		{
			name:           "returns 500 when repository list fails",
			repo:           &mockRuleRepo{listErr: errors.New("list failed")},
			wantStatusCode: http.StatusInternalServerError,
			wantListCall:   true,
			wantBody:       map[string]any{"error": "list failed"},
		},
		{
			name: "returns 200 with rules",
			repo: &mockRuleRepo{listRules: []model.Rule{
				{ID: 1, Name: "rule-a", LimitCount: 10, WindowSeconds: 60},
				{ID: 2, Name: "rule-b", LimitCount: 20, WindowSeconds: 120},
			}},
			wantStatusCode: http.StatusOK,
			wantListCall:   true,
			wantBody: []map[string]any{
				{"id": float64(1), "name": "rule-a", "limit_count": float64(10), "window_seconds": float64(60)},
				{"id": float64(2), "name": "rule-b", "limit_count": float64(20), "window_seconds": float64(120)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := newTestContext(http.MethodGet, "/rules")

			h := &RuleHandler{
				repo:    tt.repo,
				service: &mockRateLimiterService{},
			}
			h.ListRules(c)

			if w.Code != tt.wantStatusCode {
				t.Fatalf("ListRules() status code = %v, want %v", w.Code, tt.wantStatusCode)
			}
			if tt.repo.listCalled != tt.wantListCall {
				t.Fatalf("ListRules() repo called = %v, want %v", tt.repo.listCalled, tt.wantListCall)
			}

			if tt.wantStatusCode == http.StatusInternalServerError {
				var got map[string]any
				if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
					t.Fatalf("ListRules() parse error response: %v", err)
				}
				if got["error"] != tt.wantBody.(map[string]any)["error"] {
					t.Fatalf("ListRules() error body = %v, want %v", got, tt.wantBody)
				}
			}

			if tt.wantStatusCode == http.StatusOK {
				var got []map[string]any
				if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
					t.Fatalf("ListRules() parse success response: %v", err)
				}
				if len(got) != len(tt.wantBody.([]map[string]any)) {
					t.Fatalf("ListRules() result size = %v, want %v", len(got), len(tt.wantBody.([]map[string]any)))
				}
			}
		})
	}
}

func TestRuleHandler_Check(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		svc            *mockRateLimiterService
		wantStatusCode int
		wantError      bool
		wantSvcCalled  bool
		wantAllowed    bool
	}{
		{
			name:           "returns 400 when body validation fails",
			body:           `{}`,
			svc:            &mockRateLimiterService{},
			wantStatusCode: http.StatusBadRequest,
			wantError:      true,
			wantSvcCalled:  false,
		},
		{
			name:           "returns 500 when service check fails",
			body:           `{"key":"user-1","rule_id":10}`,
			svc:            &mockRateLimiterService{err: errors.New("service down")},
			wantStatusCode: http.StatusInternalServerError,
			wantError:      true,
			wantSvcCalled:  true,
		},
		{
			name:           "returns 200 with allowed true",
			body:           `{"key":"user-2","rule_id":20}`,
			svc:            &mockRateLimiterService{allowed: true},
			wantStatusCode: http.StatusOK,
			wantError:      false,
			wantSvcCalled:  true,
			wantAllowed:    true,
		},
		{
			name:           "returns 200 with allowed false",
			body:           `{"key":"user-3","rule_id":30}`,
			svc:            &mockRateLimiterService{allowed: false},
			wantStatusCode: http.StatusOK,
			wantError:      false,
			wantSvcCalled:  true,
			wantAllowed:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := newJSONContext(http.MethodPost, "/check", tt.body)

			h := &RuleHandler{
				repo:    &mockRuleRepo{},
				service: tt.svc,
			}
			h.Check(c)

			if w.Code != tt.wantStatusCode {
				t.Fatalf("Check() status code = %v, want %v", w.Code, tt.wantStatusCode)
			}

			if tt.svc.called != tt.wantSvcCalled {
				t.Fatalf("Check() service called = %v, want %v", tt.svc.called, tt.wantSvcCalled)
			}

			var body map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("Check() failed to parse response body: %v", err)
			}

			_, hasError := body["error"]
			if hasError != tt.wantError {
				t.Fatalf("Check() error field present = %v, want %v", hasError, tt.wantError)
			}

			if !tt.wantError {
				allowed, ok := body["allowed"].(bool)
				if !ok {
					t.Fatalf("Check() response did not contain boolean allowed field: %v", body)
				}
				if allowed != tt.wantAllowed {
					t.Fatalf("Check() allowed = %v, want %v", allowed, tt.wantAllowed)
				}
			}
		})
	}
}
