package io

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

type walkContext struct {
	fileChan *chan os.File
	doneChan *chan bool
	path     string
	rootFile *os.File
	debug    bool
}

//WalkDir
// Returns file chan, error, done chan
func WalkDir(filePath string, waitReadTimeout time.Duration, debug bool) (fileChan <-chan os.File, e error, doneChan <-chan bool) {
	if filePath == "" {
		return nil, errors.New("file path is nil"), nil
	}

	var file *os.File

	if file, e = os.Open(filePath); e != nil {
		return
	}

	c := &walkContext{}
	c.path = filePath
	dc := make(chan bool, 1)
	fc := make(chan os.File, 1024)
	c.doneChan = &dc
	c.fileChan = &fc
	c.rootFile = file
	c.debug = debug

	fileChan = fc
	doneChan = dc

	go func() {
		defer func() {
			*c.doneChan <- true
		}()
		defer func() {
			timeout, _ := context.WithTimeout(context.TODO(), waitReadTimeout)
			for {
				select {
				case <-timeout.Done():
					fmt.Println("Waiting fileChan consumer timeout!!! channel size: ", len(*c.fileChan))
					return
				default:
					i := len(*c.fileChan)
					if c.debug {
						fmt.Println("Least file in chan", i)
					}
					if i == 0 {
						return
					}
				}

			}
		}()
		walk(c, file)
	}()

	return
}

func walk(ctx *walkContext, file *os.File) {
	if ctx.debug {
		fmt.Printf("Read: %v\n", file.Name())
	}
	info, e := file.Stat()
	if e != nil {
		fmt.Printf("Failed to get file info of file: %v, error: %v\n", info, e.Error())
		return
	} else if !info.IsDir() {
		*ctx.fileChan <- *file
		return
	}

	infos, e2 := ioutil.ReadDir(ctx.path)
	if e2 != nil {
		fmt.Printf("Failed to list dir info of file: %v, error: %v\n", info.Name(), e2.Error())
		return
	}
	currentPath := ctx.path
	for _, infoOfChild := range infos {
		//path := fmt.Sprintf("%v/%v", currentPath, infoOfChild.Name())
		path := buildPath(currentPath, infoOfChild.Name())
		fmt.Println("Read path: ", path)
		//path := filepath.Join(currentPath, infoOfChild.Name())
		childFile, e3 := os.Open(path)
		if e3 != nil {
			fmt.Printf("Failed to list dir info of file: %v, error: %v\n", info.Name(), e3.Error())
			continue
		}
		//c := &walkContext{}
		//c.fileChan = ctx.fileChan
		//c.debug = ctx.debug
		ctx.path = path
		walk(ctx, childFile)
	}

}

func buildPath(paths ...string) string {
	if len(paths) == 0 {
		return ""
	}
	if len(paths) == 1 {
		return paths[0]
	}
	builder := strings.Builder{}
	for i, path := range paths {
		//if path == "/" {
		//	builder.WriteString(path)
		//} else {
			builder.WriteString(strings.TrimRight(path, "/"))
		//}
		if i != len(paths)-1 {
			builder.WriteString("/")
		}

	}
	return builder.String()
}
