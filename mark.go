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
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"regexp"
	"sort"
	"strconv"
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
// input Prefix is a []string list, but the output would be the string, so they need to use Join to connect
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
	// initialize the p with ""
	//for i := range p {
	//	p[i] = "\"\""
	//}
	for {
		var s string
		// fmt.Fscan reads space-separated values from an io.Reader + stops if errors occurred.
		if _, err := fmt.Fscan(br, &s); err != nil { // use &s is the requirement of the Fscan package
			break
		}
		key := p.String()
		c.chain[key] = append(c.chain[key], s)
		p.Shift(s)
	}
}

// 基于指针对象的函数
// https://docs.hacknode.org/gopl-zh/ch6/ch6-02.html

func (c *Chain) BuildFromRead(scanner *bufio.Scanner, prefixLen int) {
	p := make(map[string][]string) // We'll use this variable to hold the current prefix and mutate it with each new word we encounter.、

	// 要以 key - val的形式来储存
	// key -> string val->[]string
	// key 为前n个， val 为 后面的两个，去除掉数字的个数
	count := 0
	for scanner.Scan() {

		// 变回原来的 c.chain format
		currentLine := scanner.Text()
		//fmt.Println(currentLine)
		key, val := TextLineToChain(currentLine, prefixLen)
		p[key] = val

		// 需要一个初始化的值

		//
		count++
	}
	fmt.Println(count)
	c.chain = p
	c.prefixLen = prefixLen
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

func ValIteration(val []string) string {
	if len(val) == 1 {
		return val[0] + " 1"
	} else {
		processedVal := ""
		count := 1
		sort.Strings(val)

		for i := 0; i < len(val); i++ {
			if i < len(val)-1 && val[i] == val[i+1] {
				count++
			} else {
				//fmt.Println(count)
				processedVal = processedVal + " " + val[i] + " " + strconv.Itoa(count)
			}
		}
		return strings.TrimSpace(processedVal)
	}
}

func TextLineToChain(currentLine string, prefixLen int) (string, []string) {

	// regex
	reg := regexp.MustCompile(`\D+`)
	if reg == nil {
		fmt.Println("MustCompile err")
	}
	result := reg.FindAllString(currentLine, -1)

	// back to one string
	// []int -> string
	resultOneString := ""
	for _, s := range result {
		resultOneString = resultOneString + s
	}

	//key -> string val->[]string
	// create the format suitable for key
	splitStringList := strings.Split(resultOneString, " ")

	key := ""
	val := make([]string, 0)

	for i := 0; i < len(splitStringList)-1; i++ {
		// 前 prefixLen 作为key
		if i < prefixLen {
			if key == "\"\"" {
				key = ""
				key = key + splitStringList[i] + " "
			} else {
				key = key + splitStringList[i] + " "
			}
		} else {
			if splitStringList[i] == "" {
				fmt.Println("碰到为空的值了")
				continue
			}
			val = append(val, strings.TrimSpace(splitStringList[i]))
		}
	}

	//fmt.Println("key============", key)
	//fmt.Println("Val:", val)

	return strings.TrimSpace(key), val
}

func main() {
	// Register command-line flags => pointer. This is the default format
	mode := os.Args[1]

	if mode == "read" {
		// mode selection
		prefixLen, _ := strconv.Atoi(os.Args[2])
		outFileDir := os.Args[3]
		inFileDir := os.Args[4:]

		rand.Seed(time.Now().UnixNano()) // Seed the random number generator.

		// Initialize a new Chain.
		c := NewChain(prefixLen)

		outFile, _ := os.Create(outFileDir)
		defer outFile.Close()
		fmt.Println("We have successfully read, now the program begins:")

		count := 0
		for i := 0; i < len(inFileDir); i++ {

			// open the input files
			fi, err := os.Open(inFileDir[i])
			if err != nil {
				panic(err)
			}
			defer fi.Close()

			// Build chains from standard input.
			c.Build(fi)

			fmt.Println(c)
			// the first line, specify the number of prefix length
			if count == 0 {
				fmt.Fprintln(outFile, prefixLen)
			}
			// format: map[string][]string
			mapChain := c.chain
			//fmt.Println(mapChain)
			// key -> string val->[]string
			for key, val := range mapChain {
				//fmt.Println(key)
				fmt.Fprint(outFile, key, " ", ValIteration(val), "\n")
				//fmt.Print(key, " ", ValIteration(val), "\n")
			}
			count++
			fmt.Println("==================== one epoch finished =====================================")
		}

	} else {
		fmt.Println("Mode generate selected!!!")

		modelFileDir := os.Args[2]
		//numWords := os.Args[3]
		// 读取frequency table

		file, err := os.Open(modelFileDir)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		// read first line to gain the number of prefix
		scanner := bufio.NewScanner(file)
		numList := make([]int, 0)
		for scanner.Scan() {
			prefixLenRead, _ := strconv.Atoi(scanner.Text())
			numList = append(numList, prefixLenRead)
			break
		}

		prefixLen := numList[0]
		fmt.Println("The first line would be: ", prefixLen)

		// Reinitilize a chain
		c := NewChain(prefixLen)

		c.BuildFromRead(scanner, prefixLen)
		fmt.Println(c)

		text := c.Generate(100) // Generate text.
		fmt.Println(text)       // Write text to standard output.

	}
}
