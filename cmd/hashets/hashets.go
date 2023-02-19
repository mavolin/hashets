package main

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"flag"
	"fmt"
	"hash"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mavolin/hashets/hashets"
	"github.com/mavolin/hashets/internal/meta"
)

var (
	// FLAGS
	hashingAlgorithm hash.Hash
	ignore           []string
	replace          bool
	outPath          string
	fileNamesVar     string

	// ARGS
	inPath string

	// ENVS
	packageName string // GOPACKAGE
)

func init() {
	hashingAlgoStr := flag.String("hash", "sha256", "hashing algorithm to use (sha256, sha512, md5)")
	flag.Func("ignore", "ignores paths with the given prefix", func(s string) error {
		ignore = append(ignore, s)
		return nil
	})
	flag.BoolVar(&replace, "replace", false, "replace the original files")
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
		err := os.Mkdir(outPath, 0o755)
		if err != nil && !os.IsExist(err) {
			fmt.Fprintln(os.Stderr, "failed to create output directory:", err)
			os.Exit(1)
		}
	}

	packageName = os.Getenv("GOPACKAGE")
	if packageName == "" {
		packageName = filepath.Base(outPath)
	}
}

func usage() {
	fmt.Fprintln(flag.CommandLine.Output(), "hashets [flags] DIR")
	fmt.Fprintln(flag.CommandLine.Output())
	fmt.Fprintln(flag.CommandLine.Output(), meta.Version, "(github.com/mavolin/hashets)")
	fmt.Fprintln(flag.CommandLine.Output())
	fmt.Fprintln(flag.CommandLine.Output(), "Generate hashes for all files in DIR, and create a clone of DIR contents")
	fmt.Fprintln(flag.CommandLine.Output(), "in -o with the file names replaced with the hashes.")
	fmt.Fprintln(flag.CommandLine.Output(), "Additionally, places a file named hashets_map.go in -o, that contains")
	fmt.Fprintln(flag.CommandLine.Output(),
		"a single variable `fileNames` of type hashets.Map, which maps the original")
	fmt.Fprintln(flag.CommandLine.Output(), "file names to the hashed file names.")
	fmt.Fprintln(flag.CommandLine.Output())
	fmt.Fprintln(flag.CommandLine.Output(), "Usage:")
	flag.PrintDefaults()
}

func main() {
	m, err := hashets.HashToDir(os.DirFS(inPath), outPath, hashets.Options{
		Hash: hashingAlgorithm,
		Ignore: func(path string) bool {
			if path == "hashets_map.go" {
				return true
			}

			for _, prefix := range ignore {
				if strings.HasPrefix(path, prefix) {
					return true
				}
			}

			return false
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
