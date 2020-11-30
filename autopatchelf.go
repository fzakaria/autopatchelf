package main

import (
	"bytes"
	"debug/elf"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// Do a simple cursor selection using some program
// https://junegunn.kr/2016/02/using-fzf-in-your-program
func withFilter(command string, input func(in io.WriteCloser)) string {
	shell := os.Getenv("SHELL")
	if len(shell) == 0 {
		shell = "sh"
	}
	cmd := exec.Command(shell, "-c", command)
	cmd.Stderr = os.Stderr
	in, _ := cmd.StdinPipe()
	go func() {
		input(in)
		in.Close()
	}()
	result, _ := cmd.Output()
	return string(result)
}

func withFuzzyFilter(input func(in io.WriteCloser)) string {
	return withFilter("fzf --height=33%", input)
}

var LibraryRegex = regexp.MustCompile(`\w+`)

func normalizeLibraryName(input string) string {
	return LibraryRegex.FindString(input)
}

func determineElfDependencies(filename string) {
	file, err := elf.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v is not an ELF binary: %v\n", filename, err)
		return
	}
	fmt.Fprintf(os.Stderr, "Parsing %v\n", filename)
	libraries, err := file.ImportedLibraries()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not determine dynamic libraries for %v: %v\n", filename, err)
		return
	}

	fmt.Fprintf(os.Stderr, "Found the following libraries: %v\n", libraries)

	for _, library := range libraries {
		fmt.Fprintf(os.Stderr, "Determining /nix/store entry for %v\n", library)

		var nixLocateLine string

		out, err := exec.Command("nix-locate", "--at-root", "--whole-name", fmt.Sprintf("/lib/%v", library)).Output()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failure doing nix-locate for %v: %v\n", library, err)
			continue
		}

		// if we have no output, then an exact match was not found.
		// rely on the user to select
		if len(out) == 0 {
			fmt.Fprintf(os.Stderr, "Could not find an exact match for %v. Please select which to use.\n", library)
			normalizedLibrary := normalizeLibraryName(library)
			fmt.Fprintf(os.Stderr, "Using normalized library name: %v\n", normalizedLibrary)
			nixLocateLine = withFuzzyFilter(func(in io.WriteCloser) {
				// we no longer include --whole-name
				out, err := exec.Command("nix-locate", "--at-root", fmt.Sprintf("/lib/%v", normalizedLibrary)).Output()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failure doing nix-locate for %v\n: %v\n", normalizedLibrary, err)
					return
				}
				io.Copy(in, bytes.NewReader(out))
			})
		} else {
			nixLocateLine = withFuzzyFilter(func(in io.WriteCloser) {
				io.Copy(in, bytes.NewReader(out))
			})
		}

		fmt.Fprintf(os.Stderr, "%v has been resolved to:\n\t%v", library, nixLocateLine)

		// split the string by whitespace and the last field is the library.
		// ex.
		// glibc.out   0 s /nix/store/bpgdx6qqqzzi3szb0y3di3j3660f3wkj-glibc-2.31/lib/libc.so.6
		nixIndexOutput := strings.Fields(nixLocateLine)
		nixStorePath := nixIndexOutput[len(nixIndexOutput)-1]
		fmt.Fprintf(os.Stderr, "Realising %v\n", nixStorePath)

		cmd := exec.Command("nix-store", "--realise", nixStorePath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Failure doing nix-store realising for %v\n: %v\n", nixStorePath, err)
			return
		}
	}
}

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "You must provide a at least one binary.")
	}

	// process each argument if it's a binary
	for _, arg := range flag.Args() {
		determineElfDependencies(arg)
	}
}
