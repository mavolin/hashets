package hashets

import (
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"strings"
)

// Options provides configuration options for hashing.
type Options struct {
	// HashFunc is the hash function to use.
	//
	// Defaults to [sha256.New()].
	Hash hash.Hash

	// NamingFunc is the function used to generate the file name for the hashed
	// file.
	//
	// For each generated file, NamingFunc is called with the original file
	// name (not path) and the hash of the file.
	//
	// Defaults to DefaultNamingFunc.
	NamingFunc func(name, hash string) string

	// HashToText is the function to convert the hash to a string.
	//
	// Defaults to [hex.EncodeToString].
	HashToText func([]byte) string

	// Ignore is called for each file and, if it returns true, the file is
	// ignored, i.e. not hashed.
	//
	// Use IgnorePrefix to ignore files with a certain prefix.
	//
	// Defaults to ignoring no files.
	Ignore func(path string) bool
}

func (o *Options) setDefaults() {
	if o.Hash == nil {
		o.Hash = sha256.New()
	}

	if o.NamingFunc == nil {
		o.NamingFunc = DefaultNamingFunc
	}

	if o.HashToText == nil {
		o.HashToText = hex.EncodeToString
	}

	if o.Ignore == nil {
		o.Ignore = func(string) bool { return false }
	}
}

// IgnorePrefix returns a func to be used with [Options.Ignore] that ignores
// all files with a prefix in prefixes.
func IgnorePrefix(prefixes ...string) func(string) bool {
	return func(prefix string) bool {
		for _, p := range prefixes {
			if strings.HasPrefix(prefix, p) {
				return true
			}
		}
		return false
	}
}

// DefaultNamingFunc is the default naming function used by [Options].
//
// It generates names like "foo_1234.txt" for a file "foo.txt" with the hash
// "1234".
func DefaultNamingFunc(name, hash string) string {
	base, ext, found := strings.Cut(name, ".")
	if found {
		return base + "_" + hash + "." + ext
	}
	return base + "_" + hash
}
