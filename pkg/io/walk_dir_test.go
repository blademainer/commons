package io

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestWalkDir(t *testing.T) {
	sourceDir := "./test_source_dir"
	e := os.MkdirAll(sourceDir, os.ModePerm)
	if e != nil {
		panic(e)
	}
	if e := ioutil.WriteFile(fmt.Sprintf("%v/%v", sourceDir, "test"), []byte("hello"), os.ModePerm); e != nil {
		panic(e)
	}

	if fileChan, e, doneChan := WalkDir("./", time.Duration(10*time.Second), false); e != nil {
		panic(e)
	} else {
		for {
			select {
			case file := <-fileChan:
				fmt.Println("Found: ", file.Name())
			case <-doneChan:
				fmt.Println("Done.")
				return
			}

		}
	}
}

func ExampleWalkDir() {
	sourceDir := "./test_source_dir"
	e := os.MkdirAll(sourceDir, os.ModePerm)
	if e != nil {
		panic(e)
	}
	if e := ioutil.WriteFile(fmt.Sprintf("%v/%v", sourceDir, "test"), []byte("hello"), os.ModePerm); e != nil {
		panic(e)
	}

	if fileChan, e, doneChan := WalkDir("./", time.Duration(10*time.Second), true); e != nil {
		panic(e)
	} else {
		for {
			select {
			case file := <-fileChan:
				fmt.Println("Found: ", file.Name())
			case <-doneChan:
				fmt.Println("Done.")
				return
			}

		}
	}
}
