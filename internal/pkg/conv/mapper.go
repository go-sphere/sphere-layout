package conv

import (
	"github.com/go-viper/mapstructure/v2"
)

// Map transforms a slice of source items to a slice of target items using the provided map function.
// It applies the map function to each element and returns a new slice with the transformed results.
func Map[S any, T any](source []S, mapper func(S) T) []T {
	result := make([]T, len(source))
	for i, s := range source {
		result[i] = mapper(s)
	}
	return result
}

// MapStruct converts between struct types using mapstructure with weak decoding.
// It handles type conversions automatically and returns nil if the source is nil
// or if the conversion fails. This is useful for converting between similar struct types.
func MapStruct[S any, T any](source *S) *T {
	if source == nil {
		return nil
	}
	var target T
	err := mapstructure.WeakDecode(source, &target)
	if err != nil {
		return nil
	}
	return &target
}

// Group creates a map from a slice by extracting keys using the provided keyFunc.
// If multiple items have the same key, the last item encountered will be kept.
// This is useful for creating lookup tables from slices.
func Group[S any, K comparable](source []S, keyFunc func(S) K) map[K]S {
	result := make(map[K]S, len(source))
	for _, s := range source {
		key := keyFunc(s)
		result[key] = s
	}
	return result
}
