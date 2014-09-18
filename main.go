package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

var options = struct {
	help      *bool
	n         *bool
	recursive *bool
	v         *bool
}{
	help:      flag.Bool("help", false, "display this help and exit"),
	n:         flag.Bool("n", false, "print line number with output lines"),
	recursive: flag.Bool("r", false, "handle directories recusively"),
	v:         flag.Bool("v", false, "select non-matching lines"),
}

var (
	exitCode int
)

func match(pattern *regexp.Regexp, path string, num int, line string) {
	if !*options.v == pattern.MatchString(line) {
		reportMatch(path, num, line)
	}
}

func searchFile(pattern *regexp.Regexp, path string) error {
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	scan := bufio.NewScanner(f)
	num := 0
	for scan.Scan() {
		num++
		match(pattern, path, num, scan.Text())
	}
	return scan.Err()
}

func reportError(err error) {
	switch err := err.(type) {
	case *os.PathError:
		fmt.Fprintf(os.Stderr, "gogrep: %s: %s\n", err.Path, err.Err)
	default:
		fmt.Fprintln(os.Stderr, "gogrep:", err)
	}
	exitCode = 2
}

func reportMatch(path string, num int, line string) {
	if path != "" {
		fmt.Print(path, ":")
	}
	if *options.n {
		fmt.Print(num, ":")
	}
	fmt.Println(line)
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
		scan := bufio.NewScanner(os.Stdin)
		num := 0
		for scan.Scan() {
			num++
			match(pattern, "", num, scan.Text())
		}
		if err := scan.Err(); err != nil {
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
			}
			if info.IsDir() {
				err = filepath.Walk(name, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if info.IsDir() {
						if !*options.recursive {
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
	os.Exit(exitCode)
}
