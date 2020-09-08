package path

import (
	"path/filepath"
	"runtime"
)

// NeighborPath returns the absolute path the given relative file or directory path,
// relative to the google.golang.org/grpc/testdata directory in the user's GOPATH.
// If rel is already absolute, it is returned unmodified.
func NeighborPath(rel string) string {
	if filepath.IsAbs(rel) {
		return rel
	}
	basepath := basePath()
	return filepath.Join(basepath, rel)
}

// basePath base path of the caller
func basePath() string {
	_, currentFile, _, _ := runtime.Caller(2)
	basepath := filepath.Dir(currentFile)
	return basepath
}
