package testdata

import (
	"fmt"
	"github.com/blademainer/commons/pkg/path"
	"gotest.tools/assert"
	"path/filepath"
	"runtime"
	"testing"
)

var basepath string

func init() {
	caller, file, line, ok := runtime.Caller(0)
	fmt.Println(caller, file, line, ok)
	basepath = filepath.Dir(file)
}

func TestNeighborPath(t *testing.T) {
	s := path.NeighborPath("./a.json")
	fmt.Println(s)

	assert.Equal(t, s, filepath.Join(basepath, "/a.json"))
}
