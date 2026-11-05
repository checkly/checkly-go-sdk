package checkly

// nullIfZero returns nil for the zero value, so it is marshaled as null.
func nullIfZero[T comparable](v T) *T {
	var zero T
	if v == zero {
		return nil
	}
	return &v
}

// emptyIfNil returns an empty slice for nil, so it is marshaled as [].
func emptyIfNil[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}

func trueIfNil(b *bool) bool {
	return b == nil || *b
}
