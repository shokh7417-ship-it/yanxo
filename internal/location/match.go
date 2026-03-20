package location

// LevenshteinDistance returns edit distance between a and b (runes).
func LevenshteinDistance(a, b string) int {
	ar := []rune(a)
	br := []rune(b)
	n, m := len(ar), len(br)
	if n == 0 {
		return m
	}
	if m == 0 {
		return n
	}
	// one row + one extra slot
	prev := make([]int, m+1)
	curr := make([]int, m+1)
	for j := 0; j <= m; j++ {
		prev[j] = j
	}
	for i := 1; i <= n; i++ {
		curr[0] = i
		for j := 1; j <= m; j++ {
			cost := 1
			if ar[i-1] == br[j-1] {
				cost = 0
			}
			curr[j] = min3(curr[j-1]+1, prev[j]+1, prev[j-1]+cost)
		}
		prev, curr = curr, prev
	}
	return prev[m]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// FuzzyMatchThreshold: allow small typos (1–2 edits) depending on length.
func maxEditDistanceForFuzzy(normalizedLen int) int {
	if normalizedLen <= 3 {
		return 1
	}
	if normalizedLen <= 6 {
		return 2
	}
	return 3
}
