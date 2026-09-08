package repository

import "errors"

// ErrRuleNotFound is returned when no rule exists for the given id,
// distinguishing a missing rule from a genuine data-store failure.
var ErrRuleNotFound = errors.New("rule not found")
