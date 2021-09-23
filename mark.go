// Copyright 2011 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Generating random text: a Markov chain algorithm

Based on the program presented in the "Design and Implementation" chapter
of The Practice of Programming (Kernighan and Pike, Addison-Wesley 1999).
See also Computer Recreations, Scientific American 260, 122 - 125 (1989).

A Markov chain algorithm generates text by creating a statistical model of
potential textual suffixes for a given prefix. Consider this text:

	I am not a number! I am a free man!

Our Markov chain algorithm would arrange this text into this set of prefixes
and suffixes, or "chain": (This table assumes a prefix length of two words.)

	Prefix       Suffix

	"" ""        I
	"" I         am
	I am         a
	I am         not
	a free       man!
	am a         free
	am not       a
	a number!    I
	number! I    am
	not a        number!

To generate text using this table we select an initial prefix ("I am", for
example), choose one of the suffixes associated with that prefix at random
with probability determined by the input statistics ("a"),
and then create a new prefix by removing the first word from the prefix
and appending the suffix (making the new prefix is "am a"). Repeat this process
until we can't find any suffixes for the current prefix or we exceed the word
limit. (The word limit is necessary as the chain table may contain cycles.)

Our version of this program reads text from standard input, parsing it into a
Markov chain, and writes generated text to standard output.
The prefix and output lengths can be specified using the -prefix and -words
flags on the command-line.
*/
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"
)

// Chain contains a map ("chain") of prefixes to a list of suffixes.
// A prefix is a string of prefixLen words joined with spaces.
// A suffix is a single word. A prefix can have multiple suffixes.
type Chain struct {
	chain     map[string][]string
	prefixLen int
}

// Prefix is a Markov chain prefix of one or more words.
type Prefix []string

// String returns the Prefix as a string (string uses as a map key).
// input is a []string list, but the output would be the string, so they need to use Join to connect
func (p Prefix) String() string {
	return strings.Join(p, " ")
}

// Shift removes the first word from the Prefix and appends the given word.
//The Shift method uses the built-in copy function to copy the last len(p)-1 elements of p to the start of the slice,
// effectively moving the elements one index to the left (if you consider zero as the leftmost index).
func (p Prefix) Shift(word string) {
	copy(p, p[1:])
	p[len(p)-1] = word
}

//NewChain returns a new Chain with prefixes of prefixLen words.
//This is a constructor function
func NewChain(prefixLen int) *Chain {
	return &Chain{make(map[string][]string), prefixLen}
}

// Build reads text from the provided Reader and
// parses it into prefixes and suffixes that are stored in Chain.
// The Build method returns once the Reader's Read method returns io.EOF (end of file) or some other read error occurs.

func (c *Chain) Build(r io.Reader) {
	br := bufio.NewReader(r)       // buffering
	p := make(Prefix, c.prefixLen) // We'll use this variable to hold the current prefix and mutate it with each new word we encounter.
	for {
		var s string
		if _, err := fmt.Fscan(br, &s); err != nil { //为什么这里不能用s来代表，而必须需要用地址来代替
			break
		} // fmt.Fscan reads space-separated values from an io.Reader + stops if errors occurred.
		key := p.String()
		c.chain[key] = append(c.chain[key], s)
		p.Shift(s)
	}
}

// Generate returns a string of at most n words generated from Chain. It reads words from the map and appends them to a slice (words).
// n specifies the maximum number of integer input
func (c *Chain) Generate(n int) string {
	p := make(Prefix, c.prefixLen)
	var words []string
	for i := 0; i < n; i++ {
		choices := c.chain[p.String()]
		if len(choices) == 0 {
			break
		} // if there is not enough suffix, break the for loop
		next := choices[rand.Intn(len(choices))]
		words = append(words, next)
		p.Shift(next)
	}
	return strings.Join(words, " ")
}

func main() {
	// Register command-line flags => pointer. This is the default format
	mode := flag.String("mode", "read", "select the mode, 'read' OR 'generate'")
	numWords := flag.Int("words", 100, "maximum number of words to print") // used during reading
	prefixLen := flag.Int("prefix", 2, "prefix length in words")           // used during writing

	flag.Parse()                     // Parse command-line flags.
	rand.Seed(time.Now().UnixNano()) // Seed the random number generator.

	// Initialize a new Chain.
	c := NewChain(*prefixLen)

	// mode selection
	if *mode == "read" {
		fmt.Println("read success")

		// Build chains from standard input.
		c.Build(os.Stdin)

		// write the file
		outFile, _ := os.Create("output.txt")
		defer outFile.Close()
		fmt.Fprintln(outFile, *prefixLen)

		// format: map[string][]string
		mapChain := c.chain

		for key, val := range mapChain {
			fmt.Fprintln(outFile, key, val)
		}
	} else {
		fmt.Println("generate success")
	}

	text := c.Generate(*numWords) // Generate text.

	fmt.Println(text) // Write text to standard output.
}
