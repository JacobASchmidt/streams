package main

import (
	"fmt"

	"github.com/JacobAlbertSchmidt/streams"
)

func Alphabet() streams.Stream[rune] {
	return streams.Range('a', 'z'+1)
}
func main() {
	fmt.Println(string(streams.Collect(Alphabet()))) // produces: abcdefghijklmnopqrstuvwxyz
}
