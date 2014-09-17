package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
)

var (
	help = flag.Bool("help", false, "display this help and exit")
)

func main() {
	flag.Parse()

	if *help {
		fmt.Println("Usage: gogrep [OPTION]... PATTERN [FILE]...")
		fmt.Println("Search for PATTERN in each FILE or standard input.")
		fmt.Println("Example: gogrep -i 'hello world' menu.h main.c")
		fmt.Println()
		flag.PrintDefaults()
		return
	}

	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "Usage: gogrep [OPTION]... PATTERN [FILE]...")
		fmt.Fprintln(os.Stderr, "Try `go.grep --help' for more information.")
		os.Exit(2)
	}

	pattern, err := regexp.Compile(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	if flag.NArg() == 1 {
		scan := bufio.NewScanner(os.Stdin)
		for scan.Scan() {
			t := scan.Text()
			if pattern.MatchString(t) {
				fmt.Println(t)
			}
		}
		return
	}

	for i := 1; i < flag.NArg(); i++ {
		filename := flag.Arg(i)
		file, err := os.Open(filename)
		if err != nil {
			// test me
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		defer file.Close()
		scan := bufio.NewScanner(file)
		for scan.Scan() {
			t := scan.Text()
			if pattern.MatchString(t) {
				fmt.Println(t)
			}
		}
		if err := scan.Err(); err != nil {
			// test me
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
	}
}
