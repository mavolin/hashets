package hashets

import (
	"io/fs"
)

// FSWrapper wraps an [fs.FS] that maps hashed file names to the original file
// names of the wrapped [fs.FS].
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
