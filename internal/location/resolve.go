package location

import (
	"context"

	"yanxo/internal/repository"
)

// Resolver resolves user input to a canonical location name (alias → canonical → fuzzy).
type Resolver struct {
	repo repository.LocationRepository
}

func NewResolver(repo repository.LocationRepository) *Resolver {
	return &Resolver{repo: repo}
}

// Resolve returns the canonical location name for display/storage, or empty string if not found.
// Order: 1) alias match (normalized) 2) canonical exact normalized match 3) cautious fuzzy match.
func (r *Resolver) Resolve(ctx context.Context, userInput string) (canonical string, err error) {
	norm := Normalize(userInput)
	if norm == "" {
		return "", nil
	}

	// 1) Alias match
	canonical, err = r.repo.CanonicalByAlias(ctx, norm)
	if err != nil {
		return "", err
	}
	if canonical != "" {
		return canonical, nil
	}

	// 2) Canonical exact (normalized) match
	all, err := r.repo.AllCanonicals(ctx)
	if err != nil {
		return "", err
	}
	for _, c := range all {
		if Normalize(c) == norm {
			return c, nil
		}
	}

	// 3) Cautious fuzzy match: smallest edit distance within threshold
	maxDist := maxEditDistanceForFuzzy(len([]rune(norm)))
	var best string
	bestDist := maxDist + 1
	for _, c := range all {
		cn := Normalize(c)
		d := LevenshteinDistance(norm, cn)
		if d <= maxDist && d < bestDist {
			bestDist = d
			best = c
		}
	}
	return best, nil
}
