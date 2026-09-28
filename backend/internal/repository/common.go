package repository

import (
	"errors"

	"github.com/lib/pq"
)

// isUniqueViolation reports whether err is a PostgreSQL unique_violation
// (SQLSTATE 23505), e.g. a duplicate email, slug, or SKU.
func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	return false
}

// isForeignKeyViolation reports whether err is a PostgreSQL
// foreign_key_violation (SQLSTATE 23503).
func isForeignKeyViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23503"
	}
	return false
}
