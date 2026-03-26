package repository

import "context"

// LocationRepository provides canonical location names and aliases for search.
type LocationRepository interface {
	// CanonicalByAlias returns the canonical name for a normalized alias, or empty string.
	CanonicalByAlias(ctx context.Context, aliasNormalized string) (canonical string, err error)
	// AllCanonicals returns all canonical location names (for exact normalized match and fuzzy match).
	AllCanonicals(ctx context.Context) ([]string, error)
	// EnsureLocationWithAliases inserts canonical and its normalized aliases if not present (idempotent).
	EnsureLocationWithAliases(ctx context.Context, canonical string, aliasNormalizedList []string) error
}
