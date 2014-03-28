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

func populateLocalAudios(path string) map[string]string {
	// path -> album
	localState := make(map[string]string)

	visit := func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			xs := strings.Split(path, "/")[1:]
			album := filepath.Join(xs[0 : len(xs)-1]...)
			path := xs[len(xs)-1]
			localState[path] = album
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
			localPath := filepath.Join(path, localAlbum, v.Path)
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
			localPath := filepath.Join(path, v.Album, v.Path)
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

	api := NewVkApi(APP_ID, PERMISSIONS, API_VERSION)
	name := "id" + api.userId
	users := api.UsersGet(name)
	first := users[0]
	user := first.FirstName + " " + first.LastName +
		" (" + name + ", " + strconv.Itoa(first.Id) + ")"
	fmt.Println(users)
	fmt.Println(user)

	audioCount := api.AudioGetCount()
	fmt.Println("Number of media files:", audioCount)
	return

	albums := api.AudioGetAlbums()
	fmt.Println("Albums:", albums)

	cloudAlbumToId := make(map[string]int)
	for _, v := range albums {
		cloudAlbumToId[v.Title] = v.Id
	}

	audios := api.AudioGet(0, audioCount, &albums)
	localState := populateLocalAudios(user)
	audiosToMove := calculateAudiosToMove(user, audios, localState)
	moveAudiosToAlbums(audiosToMove, cloudAlbumToId, api)

	log.Println("Done")
}
