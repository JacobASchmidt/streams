package streams

func zero[T any]() T {
	var t T
	return t
}

func More[T any](t T) (T, bool) { return t, true }

func Done[T any]() (T, bool) { return zero[T](), false }

type Stream[T any] func() (T, bool)

func FromSlice[T any, Slice ~[]T](s Slice) Stream[T] {
	i := 0
	return func() (T, bool) {
		if i == len(s) {
			return Done[T]()
		}
		val := s[i]
		i++
		return More(val)
	}
}

func FromChan[T any, Chan ~chan T](c Chan) Stream[T] {
	return func() (T, bool) {
		val, has_val := <-c
		return val, has_val
	}
}

func Map[A, B any](in Stream[A], f func(A) B) Stream[B] {
	return func() (B, bool) {
		next, done := in()
		if done {
			return Done[B]()
		}
		return More(f(next))
	}
}

func ForEach[T any](s Stream[T], f func(T)) {
	for val, has_val := s(); has_val; val, has_val = s() {
		f(val)
	}
}

type Control int

const (
	Break Control = iota
	Continue
)

func ForEachControl[T any](s Stream[T], f func(T) Control) {
	for val, has_val := s(); has_val; val, has_val = s() {
		cntl := f(val)
		if cntl == Break {
			break
		}
	}
}

func Reduce[A, B any](s Stream[A], init B, f func(B, A) B) B {
	ForEach(s, func(a A) {
		init = f(init, a)
	})
	return init
}

func Filter[T any](s Stream[T], f func(T) bool) Stream[T] {
	return func() (T, bool) {
		val, has_val := s()
		if !has_val {
			return Done[T]()
		}
		return val, f(val)
	}
}

func Range(a, b int) Stream[int] {
	return func() (int, bool) {
		if a == b {
			return Done[int]()
		}
		next := a
		a++
		return More(next)
	}
}

func Fill[T any, Slice ~[]T](slice Slice, stream Stream[T]) {
	ForEachControl(Indices[T](slice), func(i int) Control {
		val, has_val := stream()
		if !has_val {
			return Break
		}
		slice[i] = val
		return Continue
	})
}

func Indices[T any, Slice ~[]T](s Slice) Stream[int] {
	return Range(0, len(s))
}

func Iota() Stream[int] {
	i := 0
	return func() (int, bool) {
		next := i
		i++
		return More(next)
	}
}

type Pair[A, B any] struct {
	First  A
	Second B
}

func Zip[A, B any](a Stream[A], b Stream[B]) Stream[Pair[A, B]] {
	return func() (Pair[A, B], bool) {
		next_a, has_next_a := a()
		if !has_next_a {
			return Done[Pair[A, B]]()
		}
		next_b, has_next_b := b()
		if !has_next_b {
			return Done[Pair[A, B]]()
		}
		return More(Pair[A, B]{next_a, next_b})
	}
}

func Take[T any](s Stream[T], i int) []T {
	ret := []T{}
	ForEach(Zip(s, Range(0, i)), func(p Pair[T, int]) {
		ret = append(ret, p.First)
	})
	return ret
}
