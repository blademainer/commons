package processor

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestProcessor_Process(t *testing.T) {
	sourceDir := "test_source_dir"
	targetDir := "test_target_dir"
	e := os.MkdirAll(sourceDir, os.ModePerm)
	if e != nil {
		panic(e)
	}
	for i := 0; i < 1000; i++ {
		if e := ioutil.WriteFile(fmt.Sprintf("%v/%v%v", sourceDir, "test", i), []byte("hello"), os.ModePerm); e != nil {
			panic(e)
		}
	}

	p, e := InitProcessor(sourceDir, targetDir, true, 10, true)
	if e != nil {
		panic(e)
	}

	p.Process()
}

func TestProcessor_processFile(t *testing.T) {
	sourceDir := "./test_source_dir"
	targetDir := "./test_target_dir"
	e := os.MkdirAll(sourceDir, os.ModePerm)
	if e != nil {
		panic(e)
	}
	fileName := fmt.Sprintf("%v/%v", sourceDir, "test")
	if e := ioutil.WriteFile(fileName, []byte("hello"), os.ModePerm); e != nil {
		panic(e)
	}
	processor := &Processor{}
	processor.sourcePath = sourceDir
	processor.targetPath = targetDir
	processor.debug = true
	processor.deleteSource = true
	entry := &FileEntry{}
	entry.path = fileName
	if info, e := os.Stat(fileName); e != nil {
		panic(e)
	} else {
		entry.info = info
	}
	processor.processFile(*entry)
}

func TestGetPath(t *testing.T) {
	processor := &Processor{}
	processor.sourcePath = "a/b/c"
	processor.targetPath = "c/d/e"
	subPath := "/foo/bar"
	path := processor.buildTargetPath("a/b/c" + subPath)
	fmt.Println(path)
	assert.Equal(t, path, processor.targetPath + subPath)
}

func TestGetDir(t *testing.T) {
	targetPath := "asd/bvbb"
	targetDir := filepath.Dir(targetPath)
	fmt.Println(targetDir)
}
