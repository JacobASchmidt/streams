# Streams

This package attempts to use an elegant (although potentially inefficient) approach to streams in go.

## Example
### Hello World
```go
func HelloWorld() streams.Stream[string] {
        words := make(chan string)
        go func() {
                words <- "Hello"
                words <- "To"
                words <- "The"
                words <- "World"
        }
        msgs := streams.Filter(
                streams.Receive(words),
                func (s string) bool {
                        return s == "Hello" || s == "World"
                },
        )
        streams.ForEach(
                msgs,
                func(msg string) {
                        fmt.Print(msg, " ")
                },
        )
        //prints "Hello World"
}
```

### Alphabet

```go
func Alphabet() streams.Stream[rune] {
	return streams.Range('a', 'z'+1)
}
```
### Chunk
```go
func Chunk[T any](source Stream[T], length int) Stream[[]T] {
        return func() ([]T, bool) {
                ret := Reduce(
                        Zip(Range(0, length), source), []T{}, 
                        func(ret []T, p Pair[int, T]) []T {
                                return append(ret, p.Second)
                })
                if len(ret) == 0 {
                        return Done[T]()
                }
                return More(ret)
        }
}

```

## Formalities
Below the formal definition of a stream (https://en.wikipedia.org/wiki/Stream_(computer_science))

```
Stream[t] = Done | More(t, Stream[t])
```

Note that this definition uses sum types, is recursive, and is intended for an immutable environment. Since most programming languages in industry do not have all of these characteristcs, there are a variety of approaches for implementing this data type while working around that fact.

## Survey of programming languages
### Java

Java splits up the discriminant and the value function into two methods named ```hasNext``` and ```Next```, respectively. Translating this approach to Go would look something like
```go
type Iterator[T] interface {
        HasNext() bool
        Next() T
}
```

With usage code looking something like
```go
s := iter([]int{1, 2, 3})
for s.HasNext() {
        n := s.Next()
        //use n
}
```
Note that this interface allows for obvious (and not so obvious) misuses
```go
s := iter([]int{})
s.Next() //oh no!
```

In practice, this approach allows for misuse, it a bit verbose, and makes it dificult to create new iterators

### C++
In ```C++```, there is a hierarchy of iterators, with the most basic requiring ```it != end``` for the descriminant, ```++it``` for getting the next value, and ```*it``` for getting the value out of the iterator. Translating this approach to go would look something like
```go
for it, end := Begin(slice), End(Slice); it.NotEq(end); it = it.Next() {
        n := it.Value()
        //use n
}
```

This design is the least ergonomic and has little benifit other than conformance with standard practices of early 2000's c++.

### Rust

Rust has what I will argue is the best approach of the three. 

In Rust, An iterator is an object whose next method returns an optional T. Translating this to go would look somehting like

```go
type Option[T] union {
        Some T
        None struct{}
}

type Iter[T] interface {
        Next() Option[T]
}
```

use of this would look something like
```go
it = iter(slice)
for {
        switch it {
        case None _:
                break
        case Some n:
                // use n
        }
}
```

This has the benifit of being very hard to misuse, while retating effeciency. An obvious disadvantage of this in relation to go is that native sum types don't exist in go.

## Approach

In terms of interface, I will be closely emmulating rust (as well as the formal defintion) in my design, that is, we will return either ```Done```, or ```More[T]``` from the function of our interface. These two are implemented in terms of

```go
func Done[T]() (T, bool) {
        return zero[T](), false
}
func More[T](t T) (T, bool) {
        return t, true
}
```
In that implementation, the first return value rougly corresponds to the ```value``` type seen in other languages, and the second bool type corresponds to the discriminant, or ```has-value``` type from other languages.

In dealing with "streams" from each of the above described languages, I have found that creating new streams is quite difficult. For each new operation on a stream, you often have to build a new struct that captures local variables and implements at minimum one method to fulfill the interface. 

For example, a "chunk-by" stream (which takes a streams and returns a streams of "chunks" of size n) would require quite a bit of boiler plate in any of the above approaches, looking something like this

```go
type chunker[T, Source Stream[T]] struct {
        source Source
        slice []T
}

func (c chunker[T, Source]) Next() (T, bool) {
        i := 0
        for ; i < len(slice); i++
                val, has_val := source.Next()
                if !has_val {
                        break
                }
                slice[i] = val
        }
        if i == 0 {
                return Done[T]()
        }
        return More(slice[:i])
}

func Chunk[T any, Source Stream[T]](source Source, length int) chunker[T, Source] {
        return  chuncker{source: source, slice: make([]T, length)}
}
```
This makes streams quite verbose to use, to the point where using them in production code could be quite a hassle. On top of this fact, there is a more annoying fact sometimes we want pointer receivers for streams, meaning that code even more ugly and hard to reason about. This is not ideal for extensibility.

Note in the example above (and indeed in most examples of streams), creating one of these types usually requires capturing local variables (the source and length in this case, usally a source and some sort of predicate/function), creating one member method that allows to implement the interface, and creating some type of factory function. Note that using closure, we can acheive all of these things within one simple to read function

```go

func Chunk[T any](source Stream[T], length int) Stream[[]T] {
        slice := make([]T, length)
        return func() ([]T, bool) {
                i := 0
                for ; i < len(slice); i++
                        val, has_val := source.Next()
                        if !has_val {
                                break
                        }
                        slice[i] = val
                }
                if i == 0 {
                        return Done[T]()
                }
                return More(slice[:i])
        }
}

```

or even 

```go
func Chunk[T any](source Stream[T], length int) Stream[[]T] {
        return func() ([]T, bool) {
                ret := Reduce(
                        Zip(Range(0, length), source), []T{}, 
                        func(ret []T, p Pair[int, T]) []T {
                                return append(ret, p.Second)
                })
                if len(ret) == 0 {
                        return Done[T]()
                }
                return More(ret)
        }
}

```

Note that this implementation can often be slower unless go's escape analysis works miracles (which sometimes, it does). But, all and all, it is much more extensible and easy to reason about than other similar designs.