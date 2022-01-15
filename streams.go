package streams

func zero[T any]() T {
	var t T
	return t
}

func Next[T any](t T) (T, bool) { return t, true }

func Done[T any]() (T, bool) { return zero[T](), false }

type Stream[T any] func() (T, bool)

func Map[A, B any](in Stream[A], f func(A) B) Stream[B] {
	return func() (B, bool) {
		next, done := in()
		if done {
			return Done[B]()
		}
		return Next(f(next))
	}
}
