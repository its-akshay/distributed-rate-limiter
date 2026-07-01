package repository

import (
	"context"

	"github.com/its-akshay/distributed-rate-limiter/internal/model"
)

type RuleRepositoryInterface interface {
	GetById(ctx context.Context, id int64) (*model.Rule, error)
	Create(ctx context.Context,rule *model.Rule,) error
	List(ctx context.Context) ([]model.Rule, error)
}
