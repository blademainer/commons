package processor

import (
	"context"
	"fmt"
	fio "github.com/blademainer/commons/pkg/io"
	"github.com/blademainer/commons/pkg/pool"
	"gopkg.in/go-playground/validator.v9"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Processor struct {
	sourceFile   os.File
	targetFile   os.File
	sourcePath   string
	targetPath   string
	deleteSource bool
	concurrent   int `form:"concurrent"`
	fileChan     chan FileEntry
	done         chan bool
	sync.WaitGroup
	debug bool
}

type FileEntry struct {
	info os.FileInfo
	path string
}

func InitProcessor(sourcePath string, targetPath string, deleteSource bool, concurrent int, debug bool) (p *Processor, e error) {
	validate := validator.New()
	e = validate.Var(sourcePath, "required")
	if e != nil {
		return
	}
	e = validate.Var(targetPath, "required")
	if e != nil {
		return
	}

	var sourceDir, targetDir *os.File
	if sourceDir, e = os.Open(sourcePath); e != nil {
		return
	}
	if info, e2 := sourceDir.Stat(); e2 != nil {
		panic(e2)
	} else {
		sureDir(targetPath, info.Mode())
	}
	if targetDir, e = os.Open(targetPath); e != nil {
		return
	}

	p = &Processor{}
	p.sourceFile = *sourceDir
	p.targetFile = *targetDir
	p.sourcePath = sourcePath
	p.targetPath = targetPath
	p.deleteSource = deleteSource
	p.concurrent = concurrent
	p.debug = debug
	p.fileChan = make(chan FileEntry, concurrent)
	p.done = make(chan bool, 1)
	return
}

func sureDir(path string, mode os.FileMode) {
	if _, err := os.Stat(path); err == nil {
		// path/to/whatever exists
	} else if os.IsNotExist(err) {
		err2 := os.MkdirAll(path, mode)
		if err2 != nil {
			fmt.Printf("Failed to mkdir : %v, error: %v\n", path, err2.Error())
		}
	} else {
		// Schrodinger: file may or may not exist. See err for details.

		// Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence
	}

}

func (p *Processor) waitingForWorkerDone() {
	timeout, _ := context.WithTimeout(context.TODO(), 10*time.Second)
	for {
		select {
		case <-timeout.Done():
			fmt.Println("Waiting fileChan consumer timeout!!! channel size: ", len(p.fileChan))
			return
		default:
			i := len(p.fileChan)
			if p.debug {
				fmt.Println("Least file in chan", i)
			}
			if i == 0 {
				return
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func (p *Processor) Process() {
	p.Add(2)
	go func() {
		defer func() {
			p.waitingForWorkerDone()
			p.done <- true
			p.Done()
		}()
		// waiting for done
		p.readSourceDir()
	}()
	go func() {
		defer p.Done()
		p.processTargetDir()
	}()

	p.Wait()
}

func (p *Processor) readSourceDir() {
	e := filepath.Walk(p.sourcePath,
		func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			entry := FileEntry{path: path, info: info}
			p.fileChan <- entry
			return nil
		})
	//fileChan, e, doneChan := fio.WalkDir(p.sourcePath, time.Duration(10*time.Second), true)
	if e != nil {
		panic(e)
	}
	//p.fileChan = fileChan
	fmt.Println("ReadSourceDir done.")
}

func (p *Processor) processTargetDir() {
	wp, workerDoneCh, e := pool.InitWorkerPool(p.concurrent)
	if e != nil {
		fmt.Println("Init worker with error: ", e.Error())
		panic(e)
	}

	size := 0
	doneSize := 0

	go func() {
		for {
			select {
			case <-workerDoneCh:
				fmt.Println("Done one job..")
				doneSize++
			}
		}
	}()

	for {
		select {
		case file := <-p.fileChan:
			size++
			if p.debug {
				fmt.Printf("Processing file: %v\n", file)
			}
			wp.Execute(func() error {
				p.processFile(file)
				return nil
			})
		case <-p.done:
			for i := 0; i < p.concurrent; i++ {
				if size == doneSize {
					break
				}
				time.Sleep(100 * time.Millisecond)
			}
			fmt.Println("Exit processor by done channel.")
			e := wp.Stop()
			if e != nil {
				fmt.Printf("Failed to stop workerpool! error: %v\n", e.Error())
			} else {
				fmt.Printf("Succeed stop workerpool!")
			}
			return
		}
	}
}

func (p *Processor) processFile(file FileEntry) {
	sourcePath := file.path
	targetPath := p.buildTargetPath(sourcePath)
	if p.debug {
		fmt.Printf("Processing source: %v and target: %v\n", sourcePath, targetPath)
	}
	rename := false
	var err error
	if _, err = os.Stat(targetPath); err == nil {
		// path/to/whatever exists
		sourceHash, e1 := fio.HashFileMd5(sourcePath)
		targetHash, e2 := fio.HashFileMd5(targetPath)
		if e1 != nil {
			fmt.Printf("Failed to sum of file: %v error: %v", sourcePath, e1.Error())
		} else if e2 != nil {
			fmt.Printf("Failed to sum of file: %v error: %v", targetPath, e2.Error())
		}

		if sourceHash == targetHash {
			fmt.Printf("Same file of source: %v, target: %v\n", sourcePath, targetPath)
			if p.deleteSource {
				if err := os.Remove(sourcePath); err != nil {
					fmt.Printf("Failed to remove file: %v, error: %v\n", sourcePath, err.Error())
				} else if p.debug {
					fmt.Printf("Succeed to remove file: %v\n", sourcePath)
				}
			}
		} else {
			if err := os.Rename(targetPath, targetPath+"."+targetHash); err != nil {
				fmt.Printf("Failed to rename file: %v to: %v, error: %v\n", sourcePath, targetPath, err.Error())
			} else {
				fmt.Printf("Succeed move different file of file: %v to: %v,\n", sourcePath, targetPath)
				//rename = true
			}
		}
	} else if os.IsNotExist(err) {
		rename = true
		fmt.Printf("Target file: %v is not exists so should rename from source: %v. \n", targetPath, sourcePath)
	} else {
		// Schrodinger: file may or may not exist. See err for details.

		// Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence
		fmt.Printf("Failed to stat file: %v, error: %v\n", sourcePath, err.Error())
	}
	if rename {
		sourceStat, err1 := os.Stat(sourcePath)
		if err1 != nil {
			fmt.Printf("Failed to ls stat of file: %v, error: %v\n", sourcePath, err.Error())
			return
		}

		targetDir := filepath.Dir(targetPath)
		err := os.MkdirAll(targetDir, sourceStat.Mode())
		if err != nil {
			fmt.Printf("Failed to mkdir : %v, error: %v\n", targetDir, err.Error())
		}
		// path/to/whatever does *not* exist
		if err2 := os.Rename(sourcePath, targetPath); err2 != nil {
			fmt.Printf("Failed to rename file: %v to: %v, error: %v\n", sourcePath, targetPath, err2.Error())
		} else {
			fmt.Printf("Succeed to rename file: %v to: %v\n", sourcePath, targetPath)
		}

	}
}

func (p *Processor) buildTargetPath(source string) string {
	subPath := strings.TrimPrefix(source, p.sourcePath)
	return p.targetPath + subPath
}
