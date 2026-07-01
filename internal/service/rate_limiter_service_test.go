package service

import (
	"context"
	"testing"
	"time"

	"github.com/its-akshay/distributed-rate-limiter/internal/limiter"
	"github.com/its-akshay/distributed-rate-limiter/internal/model"
	"github.com/its-akshay/distributed-rate-limiter/internal/repository"
)

type mockRepo struct {
	rule *model.Rule
	err  error
}

func (m *mockRepo) GetById(ctx context.Context, id int64) (*model.Rule, error) {
	return m.rule, m.err
}

func (m*mockRepo) Create(ctx context.Context,rule *model.Rule,) error {
	return m.err
}
func (m *mockRepo) List(ctx context.Context) ([]model.Rule, error) {
	if m.rule == nil {
		return nil, m.err
	}
	return []model.Rule{*m.rule}, m.err
}

type mockLimiter struct {
	allowed bool
	err     error
	key     string
	limit   int
	window  time.Duration
}

func (m *mockLimiter) Allow(
	ctx context.Context,
	key string,
	limit int,
	window time.Duration,
) (bool, error) {

	m.key = key
	m.limit = limit
	m.window = window
	return m.allowed, m.err
}

func TestRateLimiterService_Check(t *testing.T) {
	tests := []struct {
		name    string
		repo    repository.RuleRepositoryInterface
		limiter limiter.RateLimiter
		key     string
		ruleId  int64
		want    bool
		wantErr bool
	}{
		{
			name: "request allowed",
			repo: &mockRepo{
				rule: &model.Rule{
					ID:            1,
					LimitCount:    100,
					WindowSeconds: 60,
				},
			},
			limiter: &mockLimiter{
				allowed: true,
			},
			key:     "user123",
			ruleId:  1,
			want:    true,
			wantErr: false,
		},
		
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewRateLimiterService(tt.repo, tt.limiter)
			got, gotErr := s.Check(context.Background(), tt.key, tt.ruleId)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Check() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Check() succeeded unexpectedly")
			}
			if got != tt.want {
				t.Errorf("Check() = %v, want %v", got, tt.want)
			}
		})
	}
}
