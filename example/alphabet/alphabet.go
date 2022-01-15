package main

import (
	"fmt"

	"github.com/JacobAlbertSchmidt/streams"
)

func Alphabet() streams.Stream[rune] {
	return streams.Range('a', 'z'+1)
}

func AlphaNum() streams.Stream[rune] {
	return streams.Chain(
		Alphabet(),
		streams.Map(Alphabet(), func(r rune) rune {
			return r - 'a' + 'A'
		}),
		streams.Range('0', '9'+1),
	)
}

func main() {
	fmt.Println(string(streams.Collect(Alphabet()))) // produces: abcdefghijklmnopqrstuvwxyz
	fmt.Println(string(streams.Collect(AlphaNum())))
}
