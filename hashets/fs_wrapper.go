package hashets

import (
	"io/fs"
)

// FSWrapper wraps an [fs.FS] that maps hashed file names to the original file
// names of the wrapped [fs.FS], so that a request for "foo_1234.txt" returns
// the contents of "foo.txt".
type FSWrapper struct {
	filesys    fs.FS
	reverseMap map[string]string // hashed name -> original name
}

var (
	_ fs.FS         = (*FSWrapper)(nil)
	_ fs.ReadFileFS = (*FSWrapper)(nil)
)

// WrapFS generates file names containing hashes from the given [fs.FS] using
// Hash(filesys, o).
//
// It then returns a [FSWrapper] that, for each hashed file name, returns the
// original file.
// Those mappings are stored in the returned [Map].
//
// Files that are ignored, are left unhashed and can be accessed by their
// original file names.
func WrapFS(filesys fs.FS, o Options) (*FSWrapper, Map, error) {
	m, err := Hash(filesys, o)
	if err != nil {
		return nil, nil, err
	}

	reverseMap := make(map[string]string, len(m))
	for k, v := range m {
		reverseMap[v] = k
	}

	return &FSWrapper{
		filesys:    filesys,
		reverseMap: reverseMap,
	}, m, nil
}

// Open returns the file represented by the passed hashed name.
//
// If there is no file mapped to the passed name, it looks for directly for a
// file with the given name.
func (fsw *FSWrapper) Open(name string) (fs.File, error) {
	if origName, ok := fsw.reverseMap[name]; ok {
		name = origName
	}

	return fsw.filesys.Open(name)
}

func (fsw *FSWrapper) ReadFile(name string) ([]byte, error) {
	if origName, ok := fsw.reverseMap[name]; ok {
		name = origName
	}

	return fs.ReadFile(fsw.filesys, name)
}
