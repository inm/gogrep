package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

var options = struct {
	help *bool
	n    *bool
	q    *bool
	r    *bool
	s    *bool
	v    *bool
}{
	help: flag.Bool("help", false, "display this help and exit"),
	n:    flag.Bool("n", false, "print line number with output lines"),
	q:    flag.Bool("q", false, "suppress all normal output"),
	r:    flag.Bool("r", false, "handle directories recusively"),
	s:    flag.Bool("s", false, "suppress error messages"),
	v:    flag.Bool("v", false, "select non-matching lines"),
}

var (
	exitCode   int
	matchCount int
)

func reportError(err error) {
	if !*options.s {
		switch err := err.(type) {
		case *os.PathError:
			fmt.Fprintf(os.Stderr, "gogrep: %s: %s\n", err.Path, err.Err)
		default:
			fmt.Fprintln(os.Stderr, "gogrep:", err)
		}
	}
	exitCode = 2
}

func searchFile(pattern *regexp.Regexp, path string) error {
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	return search(pattern, path, f)
}

func search(pattern *regexp.Regexp, path string, r io.Reader) error {
	scan := bufio.NewScanner(r)
	num := 0
	for scan.Scan() {
		num++
		line := scan.Text()
		if !*options.v == pattern.MatchString(line) {
			matchCount++
			if *options.q {
				continue
			}
			fmt.Print(path, ":")
			if *options.n {
				fmt.Print(num, ":")
			}
			fmt.Println(line)
		}
	}
	return scan.Err()
}

func main() {
	flag.Parse()

	if *options.help {
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
		if err := search(pattern, "", os.Stdin); err != nil {
			reportError(err)
		}
		return
	}

	for i := 1; i < flag.NArg(); i++ {
		filenames, err := filepath.Glob(flag.Arg(i))
		if err != nil {
			reportError(err)
		}
		for _, name := range filenames {
			info, err := os.Stat(name)
			if err != nil {
				reportError(err)
				continue
			}
			if info.IsDir() {
				err = filepath.Walk(name, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if info.IsDir() {
						if !*options.r {
							return filepath.SkipDir
						}
						return nil
					}
					return searchFile(pattern, path)
				})
				if err != nil {
					reportError(err)
				}
				continue
			}
			if err := searchFile(pattern, name); err != nil {
				reportError(err)
			}
		}
	}
	if *options.q {
		if exitCode == 0 && matchCount == 0 {
			os.Exit(1)
		}
	}
	os.Exit(exitCode)
}
