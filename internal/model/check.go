package model

type CheckRequest struct {
	Key    string `json:"key" binding:"required"`
	RuleID int64  `json:"rule_id" binding:"required"`
}

type CheckResponse struct {
	Allowed bool `json:"allowed"`
}

