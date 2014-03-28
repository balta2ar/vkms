package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func urlSize(url string) int64 {
	resp, err := http.Head(url)
	if err != nil {
		log.Println("Cannot download url", url)
		return -1
	}

	return resp.ContentLength
}

func fileSize(path string) int64 {
	fi, err := os.Stat(path)
	if err != nil {
		log.Println("Cannot stat path", path, err)
		return -1
	}
	return fi.Size()
}

func download(url string, path string) {
	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		log.Println("Cannot create dir for path", path)
	}

	out, err := os.Create(path)
	if err != nil {
		log.Println("Cannot create path", path)
		return
	}
	defer out.Close()
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Cannot download url", url)
		return
	}
	defer resp.Body.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Println("Problems while saving to disk", err)
	}
}

func exists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}
