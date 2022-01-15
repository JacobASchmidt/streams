# Streams

This package attempts to use an elegant (although potentially inefficient) approach to streams in go.

## Yet another Iterator implementation, Why?

This package takes a novel approach to iterators which allows for more clear and compact code as opposed to other apporaches. Other approaches create a unique type for every adapter, this approach uses closures. Look at alphabet raw example for a clear example of this distinction.
## Example
### Hello World (slice)
```go
func HelloWorld() streams.Stream[string] {
        words := []string{"Hello", "To", "The", "World"}
        msgs := streams.Filter(
                streams.Elements(words),
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

### Hello World (channels)
```go
func HelloWorld() streams.Stream[string] {
        words := make(chan string)
        go func() {
                words <- "Hello"
                words <- "To"
                words <- "The"
                words <- "World"
                close(words)
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

### Alphabet (raw implementation)
```go
func Alphabet() streams.Stream[rune] {
        r := 'a'
        return func() (rune, bool) {
                if r > 'z' {
                        return streams.Done[rune]()
                }
                next := r
                r++
                return streams.More(next)
        }
}
```

### Chunk
```go
func Chunk[T any](source Stream[T], length int) Stream[[]T] {
        return func() ([]T, bool) {
                ret := streams.Reduce(
                        streams.Zip(Range(0, length), source), []T{}, 
                        func(ret []T, p Pair[int, T]) []T {
                                return append(ret, p.Second)
                })
                if len(ret) == 0 {
                        return streams.Done[T]()
                }
                return streams.More(ret)
        }
}

```

## Formalities
Below the formal definition of a stream (https://en.wikipedia.org/wiki/Stream_(computer_science))

```
Stream[t] = Done | More(t, Stream[t])
```

Note that this definition uses sum types, is recursive, and is intended for an immutable environment. Since most programming languages in industry do not have all of these characteristcs, there are a variety of approaches for implementing this data type while working around that fact.

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

I will also use a quite unique approach of a stream of T being just a function that returns either Done or More, instead of an interface. This allows it to be easier to write custom adaptors.

## Why use function instead of interface?

In dealing with "streams" from each of the below described languages, I have found that creating new streams is quite difficult. For each new operation on a stream, you often have to build a new struct that captures local variables and implements at minimum one method to fulfill the interface. 

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
Using functions instead leaves just

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
I think there are clear advantages to function approach as compared to the interface approach (clear return type, clear input type, automatic captures, etc).

Note that this implementation can often be slower unless go's escape analysis works miracles (which sometimes, it does). But, all and all, it is much more extensible and easy to reason about than other similar designs.


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
