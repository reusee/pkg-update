package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

var (
	pt = fmt.Printf
	sp = fmt.Sprintf

	parallel int
)

func init() {
	flag.IntVar(&parallel, "j", 8, "the number of parallel commands")
	flag.Parse()
}

func main() {
	t0 := time.Now()

	checkErr := func(msg string, err error) {
		if err != nil {
			log.Fatalf("%s: %v", msg, err)
		}
	}

	path := os.Getenv("GOPATH")
	path, err := filepath.Abs(path)
	checkErr("get absolute path of GOPATH", err)

	dirs := []string{}
	filepath.Walk(filepath.Join(path, "src"), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			return nil
		}
		base := filepath.Base(path)
		if base == ".git" {
			dir := filepath.Dir(path)
			dirs = append(dirs, dir)
			return filepath.SkipDir
		}
		return nil
	})

	printer := make(chan string)
	go func() {
		for s := range printer {
			print(s)
		}
	}()

	wg := new(sync.WaitGroup)
	wg.Add(len(dirs))
	sem := make(chan struct{}, parallel)
	for _, dir := range dirs {
		sem <- struct{}{}
		dir := dir
		go func() {
			defer func() {
				wg.Done()
				<-sem
			}()
			err = os.Chdir(dir)
			checkErr(sp("change dir to %s", path), err)
			output, err := exec.Command("git", "pull").CombinedOutput()
			checkErr(sp("run git pull in dir %s", dir), err)
			printer <- sp("%s: %s", dir, output)
		}()
	}
	wg.Wait()

	pt("update %d packages in %v\n", len(dirs), time.Now().Sub(t0))
}
