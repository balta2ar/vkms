package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	APP_ID      = "4263245"
	PERMISSIONS = "audio"
	API_VERSION = "5.16"
)

func getAlbumFromPath(path string, user string) string {
	i := strings.Index(path, user)
	p := i + len(user) + 1
	rest := path[p:]

	xs := strings.Split(rest, "/")
	if len(xs) > 1 {
		return filepath.Join(xs[0 : len(xs)-1]...)
	} else {
		return ""
	}
}

func populateLocalAudios(path string, user string) map[string]string {
	// path -> album
	localState := make(map[string]string)

	visit := func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return nil
		}
		if !f.IsDir() {
			localState[path] = getAlbumFromPath(path, user)
		}
		return nil
	}

	filepath.Walk(path, visit)
	return localState
}

func calculateAudiosToMove(
	path string, audios []Audio,
	localState map[string]string) map[string][]string {

	// album -> []id
	audiosToMove := make(map[string][]string)
	for _, v := range audios {
		localAlbum, ok := localState[v.Path]
		if ok {
			// we have file with such name locally
			localPath := v.Path
			if urlSize(v.Url) != fileSize(localPath) {
				log.Println("different size, download", v.Title)
				download(v.Url, localPath)
				continue
			}

			if v.Album == localAlbum {
				log.Println("present locally, same album", v.Title)
			} else {
				log.Println("present locally, different album, moving: local", localAlbum, "cloud", v.Album, v.Title)
				if _, ok := audiosToMove[localAlbum]; !ok {
					audiosToMove[localAlbum] = make([]string, 0)
				}
				audiosToMove[localAlbum] = append(audiosToMove[localAlbum], strconv.Itoa((v.Id)))
			}
		} else {
			// file is not present locally
			localPath := v.Path
			log.Println("missing locally, download", v.Title)
			download(v.Url, localPath)
		}
	}
	fmt.Println(audiosToMove)
	return audiosToMove
}

func moveAudiosToAlbums(
	audiosToMove map[string][]string,
	cloudAlbumToId map[string]int,
	api *VkApi) {

	for albumTitle, audioIds := range audiosToMove {
		localAlbumId, ok := cloudAlbumToId[albumTitle]
		if !ok && albumTitle != "" {
			// create if it does not exist and its not empty
			localAlbumId := api.AudioAddAlbum(albumTitle)
			if localAlbumId != 0 {
				log.Println("Added new album, id", localAlbumId, albumTitle)
			} else {
				log.Println("Could not add album", albumTitle)
				continue
			}
		}
		ids := strings.Join(audioIds, ",")
		if !api.AudioMoveToAlbum(localAlbumId, ids) {
			log.Fatalln("Could not move to album", localAlbumId, ids)
		}
	}
}

func main() {
	log.Println("Hello")
	base := "."
	if len(os.Args) > 1 {
		base = os.Args[1]
	}

	var api *VkApi = nil
	api = NewVkApi(APP_ID, PERMISSIONS, API_VERSION)

	// make sure we have valid token first
	valid := false
	for i := 0; i < 3; i++ {
		count, err := api.AudioGetCount()
		log.Println("AudioGetCount:", i, count, err)
		if err == nil {
			log.Println("Count", count, "Token received:", api.Token)
			valid = true
			break
		}
	}
	if !valid {
		log.Fatalln("Failed to get a valid token")
	}

	api.saveToken(TOKEN_FILENAME)
	audioCount, _ := api.AudioGetCount()
	log.Println("Number of media files:", audioCount)

	albums := api.AudioGetAlbums()
	log.Println("Albums:", albums)

	cloudAlbumToId := make(map[string]int)
	for _, v := range albums {
		cloudAlbumToId[v.Title] = v.Id
	}

	path := filepath.Join(base, api.User)
	audios := api.AudioGet(0, audioCount, &albums, path)
	localState := populateLocalAudios(path, api.User)
	audiosToMove := calculateAudiosToMove(path, audios, localState)
	moveAudiosToAlbums(audiosToMove, cloudAlbumToId, api)

	log.Println("Done")
}
