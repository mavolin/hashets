// Package hashets provides hash-based cache busting for [fs.FS].
package hashets

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Map represents a map of file paths to file paths with hashed names.
//
// For example for a file ./foo.txt with the hash 1234.
// Map["foo.txt"] would return "foo_1234.txt", which is the name of the hashed
// file.
type Map map[string]string

// Get returns the hashed file path for the given file path to the unhashed
// equivalent.
//
// If m is nil, the original file path is returned.
// This is useful if the unhashed files are replaced with hashed files in
// production, but are left unhashed in development.
//
// See the readme of this repository for an example of such a case.
func (m Map) Get(name string) string {
	if m == nil {
		return name
	}

	return m[name]
}

// Hash takes the given [fs.FS], hashes all its files using the options
// provided, and returns a [Map] that maps the original file path to the
// same path, but with the file name replaced with the hashed file name, as
// returned by [Options.NamingFunc].
func Hash(inFS fs.FS, o Options) (Map, error) {
	o.setDefaults()

	m := make(Map)
	err := fs.WalkDir(inFS, ".", func(p string, dir fs.DirEntry, _ error) error {
		p = strings.TrimPrefix(p, "./")
		if dir == nil || dir.IsDir() {
			return nil
		}

		if o.Ignore(p) {
			return nil
		}

		in, err := inFS.Open(p)
		if err != nil {
			return err
		}

		o.Hash.Reset()
		_, err = io.Copy(o.Hash, in)
		if err != nil {
			return err
		}

		if err := in.Close(); err != nil {
			return err
		}

		name := path.Base(p)
		hash := o.Hash.Sum(nil)
		m[p] = p[:len(p)-len(name)] + o.NamingFunc(name, o.HashToText(hash))
		return nil
	})
	if err != nil {
		return nil, err
	}

	return m, nil
}

// HashFile takes the given [io.Reader], calculates the hash of its contents
// and returns the hashed file name, as returned by [Options.NamingFunc], using
// name as the original file name.
func HashFile(name string, in io.Reader, o Options) (string, error) {
	o.setDefaults()

	o.Hash.Reset()
	if _, err := io.Copy(o.Hash, in); err != nil {
		return "", err
	}

	hash := o.Hash.Sum(nil)
	return o.NamingFunc(name, o.HashToText(hash)), nil
}

// HashToTempDir takes the given [fs.FS], hashes all its files using the options
// provided and returns a new [fs.FS] that stores the hashed files in a temp
// directory on the local filesystem.
//
// If HashToTempDir returns without an error, the temporary directory has been
// created, and it is the responsibility of the caller to call the returned
// cleanup function to remove the temporary directory after it is no longer
// needed, most commonly after the program exits.
//
// If HashToTempDir has created the temporary directory, but returns an error,
// the temporary directory will have been removed by HashToTempDir itself, and
// the caller need not call the cleanup function.
// However, the cleanup will never be nil, so it is safe to call it even if
// an error was returned.
//
// The returned [Map] provides mappings from the original file path to the same
// path, but with the file name replaced with the hashed file name, as returned
// by [Options.NamingFunc].
func HashToTempDir(inFS fs.FS, o Options) (_ fs.FS, _ Map, cleanup func() error, _ error) {
	cleanup = func() error { return nil }

	outPath, err := os.MkdirTemp("", "hashets")
	if err != nil {
		return nil, nil, cleanup, err
	}

	cleanup = func() error { return os.RemoveAll(outPath) }

	m, err := HashToDir(inFS, outPath, o)
	if err != nil {
		_ = cleanup()
		return nil, nil, func() error { return nil }, err
	}

	return os.DirFS(outPath), m, cleanup, nil
}

// HashToDir takes the given [fs.FS], hashes all its files using the options
// provided and writes the hashed files to the given directory.
//
// The returned [Map] provides mappings from the original file path to the same
// path, but with the file name replaced with the hashed file name, as returned
// by [Options.NamingFunc].
//
// It is explicitly allowed for the output directory to match the input [fs.FS].
// That means HashToDir(os.DirFS("/some/path"), "/some/path", o) is valid and
// will work as expected.
func HashToDir(inFS fs.FS, outPath string, o Options) (Map, error) {
	o.setDefaults()

	m := make(Map)
	err := fs.WalkDir(inFS, ".", func(path string, dir fs.DirEntry, err error) error {
		path = strings.TrimPrefix(path, "./")
		if dir == nil || dir.IsDir() {
			if dir != nil && dir.IsDir() && path != "." {
				if err := os.Mkdir(outPath+"/"+path, 0o755); err != nil {
					if !errors.Is(err, os.ErrExist) {
						return err
					}
				}
			}

			return nil
		}

		if o.Ignore(path) {
			return nil
		}

		return writeHashed(inFS, path, outPath, m, o)
	})
	if err != nil {
		return nil, err
	}

	return m, nil
}

// ============================================================================
// Utils
// ======================================================================================

func writeHashed(inFS fs.FS, inPath, outPath string, m Map, o Options) error {
	in, err := inFS.Open(inPath)
	if err != nil {
		return err
	}

	o.Hash.Reset()
	_, err = io.Copy(o.Hash, in)
	if err != nil {
		return err
	}

	if err := in.Close(); err != nil {
		return err
	}

	name := path.Base(inPath)
	hash := o.Hash.Sum(nil)
	hashedPath := inPath[:len(inPath)-len(name)] + o.NamingFunc(name, o.HashToText(hash))
	m[inPath] = hashedPath

	in, err = inFS.Open(inPath)
	if err != nil {
		return err
	}

	stat, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(filepath.Join(outPath, hashedPath), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, stat.Mode()&0o555)
	if err != nil {
		return err
	}

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	if err := in.Close(); err != nil {
		return err
	}

	if err := out.Close(); err != nil {
		return err
	}

	return nil
}
