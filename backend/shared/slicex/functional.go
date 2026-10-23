package slicex

func Map[T any, R any](items []T, fn func(T) R) []R {
	res := make([]R, len(items))
	for i, v := range items {
		res[i] = fn(v)
	}
	return res
}

func Filter[T any](items []T, fn func(T) bool) []T {
	res := make([]T, 0, len(items))
	for _, v := range items {
		if fn(v) {
			res = append(res, v)
		}
	}
	return append([]T{}, res...)
}

func Find[T any](items []T, fn func(T) bool) (T, bool) {
	for _, v := range items {
		if fn(v) {
			return v, true
		}
	}
	var zero T
	return zero, false
}
