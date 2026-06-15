package repository

import (
	"context"

	"github.com/its-akshay/distributed-rate-limiter/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RuleRepository struct {
	db *pgxpool.Pool
}

func NewRuleRepository(db *pgxpool.Pool) *RuleRepository {
	return &RuleRepository{
		db: db,
	}
}

func (r *RuleRepository) Create(
	ctx context.Context,
	rule *model.Rule,
) error {
	query := `
	INSERT INTO rules(name, limit_count, window_seconds)
	VALUES ($1, $2, $3)
	RETURNING id, created_at
	`
	return r.db.QueryRow(
		ctx,
		query,
		rule.Name,
		rule.LimitCount,
		rule.WindowSeconds,
	).Scan(&rule.ID, &rule.CreatedAt)
}

func (r *RuleRepository) GetById(ctx context.Context, id int64) (*model.Rule, error) {
	query := `
	SELECT id, name, limit_count, window_seconds, created_at
	FROM rules
	WHERE id = $1
	`
	var rule model.Rule
	err := r.db.QueryRow(ctx, query, id).Scan(
		&rule.ID,
		&rule.Name,
		&rule.LimitCount,
		&rule.WindowSeconds,
		&rule.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &rule, nil

}

func (r *RuleRepository) List(ctx context.Context) ([]model.Rule, error) {
	query := `
	SELECT id, name, limit_count, window_seconds, created_at
	FROM rules
	ORDER BY id
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var rules []model.Rule

	for rows.Next() {
		var rule model.Rule

		err := rows.Scan(
			&rule.ID,
			&rule.Name,
			&rule.LimitCount,
			&rule.WindowSeconds,
			&rule.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}
