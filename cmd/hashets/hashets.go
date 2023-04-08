// Command hashets provides the CLI for hashets.
// It can be used to create hashes for files and directories.
package main

import (
	"crypto/md5" //nolint:gosec
	"crypto/sha256"
	"crypto/sha512"
	"flag"
	"fmt"
	"hash"
	"os"
	"path/filepath"
	"sort"

	"github.com/bmatcuk/doublestar"

	"github.com/mavolin/hashets/hashets"
	"github.com/mavolin/hashets/internal/meta"
)

var (
	//
	// FLAGS

	hashingAlgorithm hash.Hash
	ignore           []string
	include          []string
	replace          bool
	outPath          string
	fileNamesVar     string

	//
	// ARGS.

	inPath string

	//
	// ENVS.

	packageName string // GOPACKAGE
)

func init() {
	hashingAlgoStr := flag.String("hash", "sha256", "hashing algorithm to use (sha256, sha512, md5)")
	flag.Func("ignore",
		"ignores paths that match the glob\n"+
			"supports ** globs",
		func(s string) error {
			_, err := doublestar.PathMatch(s, "")
			if err != nil {
				return err
			}

			ignore = append(ignore, s)
			return nil
		})
	flag.Func("include",
		"includes only paths that match the glob\n"+
			"if both -include and -ignore are set, a file must be included and not ignored to be hashed\n"+
			"supports ** globs",
		func(s string) error {
			_, err := doublestar.PathMatch(s, "")
			if err != nil {
				return err
			}

			include = append(include, s)
			return nil
		})
	flag.BoolVar(&replace, "replace", false, "delete the original original files after hashing")
	flag.StringVar(&outPath, "o", "", "output directory (default DIR)")
	flag.StringVar(&fileNamesVar, "var", "FileNames", "name of the variable in hashets_map.go")

	flag.CommandLine.Usage = usage
	flag.Parse()

	switch *hashingAlgoStr {
	case "sha256":
		hashingAlgorithm = sha256.New()
	case "sha512":
		hashingAlgorithm = sha512.New()
	case "md5":
		hashingAlgorithm = md5.New() //nolint:gosec
	default:
		fmt.Fprintln(os.Stderr, "invalid hashing algorithm:", *hashingAlgoStr)
		os.Exit(1)
	}

	if len(flag.Args()) != 1 {
		flag.CommandLine.Usage()
		os.Exit(1)
	}

	inPath = filepath.Clean(flag.Arg(0))
	if outPath == "" {
		outPath = inPath
	} else {
		outPath = filepath.Clean(outPath)
		err := os.Mkdir(outPath, 0o755)
		if err != nil && !os.IsExist(err) {
			fmt.Fprintln(os.Stderr, "failed to create output directory:", err)
			os.Exit(1)
		}
	}

	if filepath.Clean(outPath) == "." {
		packageName = os.Getenv("GOPACKAGE")
		if packageName == "" {
			abs, err := filepath.Abs(outPath)
			if err != nil {
				fmt.Fprintln(os.Stderr, "failed to get absolute path:", err)
				os.Exit(1)
			}

			packageName = filepath.Base(abs)
		}
	} else {
		packageName = filepath.Base(outPath)
		if packageName == "." {
			abs, err := filepath.Abs(outPath)
			if err != nil {
				fmt.Fprintln(os.Stderr, "failed to get absolute path:", err)
				os.Exit(1)
			}

			packageName = filepath.Base(abs)
		}
	}
}

func usage() {
	fmt.Fprintln(flag.CommandLine.Output())
	fmt.Fprintln(flag.CommandLine.Output(), meta.Version, "(github.com/mavolin/hashets)")
	fmt.Fprintln(flag.CommandLine.Output())
	fmt.Fprintln(flag.CommandLine.Output(), "Generate hashes for all files in DIR, and create a clone of DIR's contents")
	fmt.Fprintln(flag.CommandLine.Output(), "in -o with the file names including hashes hashes.")
	fmt.Fprintln(flag.CommandLine.Output(), "Additionally, places a file named hashets_map.go in -o, that contains")
	fmt.Fprintln(flag.CommandLine.Output(), "a single variable `FileNames` of type hashets.Map, which maps the original")
	fmt.Fprintln(flag.CommandLine.Output(), "file names to the hashed file names.")
	fmt.Fprintln(flag.CommandLine.Output())
	fmt.Fprintln(flag.CommandLine.Output(), "Usage:")
	fmt.Fprintln(flag.CommandLine.Output(), " hashets [flags] DIR")
	flag.PrintDefaults()
}

func main() {
	m, err := hashets.HashToDir(os.DirFS(inPath), outPath, hashets.Options{
		Hash: hashingAlgorithm,
		Ignore: func(p string) bool {
			if p == "hashets_map.go" {
				return true
			}

			for _, pattern := range ignore {
				// filepath.Clean to convert the slash-based fs path to an
				// os-style path
				if match, _ := doublestar.PathMatch(pattern, filepath.Clean(p)); match {
					return true
				}
			}

			if len(include) == 0 {
				return false
			}

			for _, pattern := range include {
				if match, _ := doublestar.PathMatch(pattern, filepath.Clean(p)); match {
					return false
				}
			}

			return true
		},
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to hash files:", err)
		os.Exit(1)
	}

	if replace {
		for origName := range m {
			if err := os.Remove(filepath.Join(outPath, origName)); err != nil {
				fmt.Fprintln(os.Stderr, "replace: failed to remove original file:", err)
				os.Exit(1)
			}
		}
	}

	mapFile, err := os.Create(filepath.Join(outPath, "hashets_map.go"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to create map file:", err)
		os.Exit(1)
	}

	fmt.Fprintln(mapFile, "package", packageName)
	fmt.Fprintln(mapFile)
	fmt.Fprintln(mapFile, `import "github.com/mavolin/hashets/hashets"`)
	fmt.Fprintln(mapFile)
	fmt.Fprintln(mapFile, "// Code generated by hashets. DO NOT EDIT.")
	fmt.Fprintln(mapFile)
	fmt.Fprintln(mapFile, "var", fileNamesVar, "= hashets.Map{")

	// so that two runs of hashets with the same input produce the same output
	names := make([]string, 0, len(m))
	for name := range m {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		fmt.Fprintf(mapFile, "\t%q: %q,\n", name, m[name])
	}

	fmt.Fprintln(mapFile, "}")
}
