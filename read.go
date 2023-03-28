package main

import (
	"bufio"
	"os"
	"path"
	"path/filepath"
)

func Read(s chan string) {
	defer close(s)
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	wd = path.Join(wd, "log")
	// 可以通过递归调用实现获取所有子目录的文件名
	_ = filepath.Walk(wd, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			s <- scanner.Text()
		}
		return nil
	})

}
