package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	E_TOKEN_EXPIRED = 5
	TOKEN_FILENAME  = "token.json"
)

type VkApi struct {
	Token   string
	UserId  string
	Version string
	User    string
}

type ResponseGetAudio struct {
	Response struct {
		Count float64
		Items []Audio
	}
}
type Audio struct {
	Artist   string
	Title    string
	Url      string
	Album    string
	AlbumId  int `json:"album_id"`
	Path     string
	Id       int
	Duration int
	GenreId  int `json:"genre_id"`
	OwnerId  int `json:"owner_id"`
}

type ResponseUsersGet struct {
	Response []User
}

type User struct {
	Id        int
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type ResponseAudioAddAlbum struct {
	Response struct {
		AlbumId int `json:"album_id"`
	}
}

type ResponseInt struct {
	Response int
}

type Album struct {
	Id      int
	OwnerId int `json:"owner_id"`
	Title   string
}

type ResponseError struct {
	Error *struct {
		ErrorCode int    `json:"error_code"`
		ErrorMsg  string `json:"error_msg"`
	}
}

type ResponseAlbums struct {
	Response struct {
		Count float64
		Items []Album
	}
}

func getUrl(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Cannot get audio")
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Cannot read body")
		return nil, err
	}
	ioutil.WriteFile("body.txt", body, 0600)
	return body, nil
}

func (api *VkApi) handleError(response ResponseError) bool {
	if response.Error.ErrorCode == E_TOKEN_EXPIRED {
		token, userId, err := getNewToken(APP_ID, PERMISSIONS, API_VERSION)
		if err == nil {
			log.Println("handleError: new token:", token, userId)
			api.Token = token
			api.UserId = userId
			return true
		} else {
			log.Fatalln("handleError: Cannot get new token")
		}
	} else {
		log.Println("Cannot handle this error")
	}
	return false
}

func (api *VkApi) request(url string, obj interface{}) error {
	body, err := getUrl(url)
	if err != nil {
		log.Fatal("Cannot get url", url)
	}
	var response ResponseError
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Println(err)
		log.Fatalln("Cannot parse response as Error")
	}
	if response.Error != nil {
		log.Println("Error field is set")
		log.Println(*response.Error)
		if api.handleError(response) {
			log.Println("Error has been handled")
			return errors.New("Error has been handled")
		}
		log.Fatalln("Error is set and cannot be handled, cannot continue")
	}
	err = json.Unmarshal(body, obj)
	if err != nil {
		log.Println(err)
		log.Fatalln("Cannot parse response as ConcreteType")
	}
	return nil
}

func input(msg string) (string, error) {
	log.Printf("Paste here the URL that you were redirected to:")
	in := bufio.NewReader(os.Stdin)
	return in.ReadString('\n')
}

func getNewToken(appId string, permissions string, version string) (token string, userId string, err error) {

	authUrl := "https://oauth.vk.com/authorize" +
		"?client_id=" + appId +
		"&scope=" + permissions +
		"&redirect_uri=https://oauth.vk.com/blank.html" +
		"&display=page" +
		"&v=" + version +
		"&response_type=token"

	log.Printf("Please visit this URL to authorize the appplication: %s\n", authUrl)
	line, err := input(authUrl)
	if err != nil {
		log.Fatalln("Cannot get line from stdin")
	}
	log.Println("line:", line)

	origAccessUrl := strings.Trim(line, "\r\n")
	accessUrl := strings.Replace(origAccessUrl, "#", "?", -1)
	parsedUrl, err := url.Parse(accessUrl)
	if err != nil {
		log.Println("Cannot parse accessUrl:", accessUrl)
		return "", "", err
	}
	fmt.Println(parsedUrl)

	token = parsedUrl.Query().Get("access_token")
	userId = parsedUrl.Query().Get("user_id")
	log.Println("token:", token, "user_id", userId)
	return
}

func loadToken(path string) (token string, userId string, err error) {
	log.Println("Loading token from file", path)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println("Cannot read file", path, err)
		return "", "", err
	}
	var fakeApi VkApi
	err = json.Unmarshal(data, &fakeApi)
	return fakeApi.Token, fakeApi.UserId, err
}

func (api *VkApi) saveToken(path string) {
	log.Println("Saving token to file", path, *api)
	data, err := json.Marshal(api)
	if err != nil {
		log.Println("saveToken:", err)
		return
	}
	err = ioutil.WriteFile(path, data, 0600)
	if err == nil {
		log.Println("Saved token to file", path)
	} else {
		log.Println("Could not save token to file", path)
	}
}

func NewVkApi(appId string, permissions string, version string) *VkApi {
	token, userId, err := loadToken(TOKEN_FILENAME)
	log.Println(token, userId, err)

	if err != nil {
		token, userId, err = getNewToken(appId, permissions, version)
	}

	api := &VkApi{token, userId, version, ""}

	name := "id" + api.UserId
	users := api.UsersGet(name)
	first := users[0]
	user := first.FirstName + " " + first.LastName +
		" (" + name + ", " + strconv.Itoa(first.Id) + ")"
	api.User = user

	return api
}

func (api *VkApi) UsersGet(ids string) []User {
	var response ResponseUsersGet
	api.request("https://api.vk.com/method/users.get"+
		"?user_ids="+ids+
		"&v="+api.Version, &response)
	return response.Response
}

func (api *VkApi) AudioGetCount() (int, error) {
	var response ResponseInt
	err := api.request("https://api.vk.com/method/audio.getCount"+
		"?access_token="+api.Token+
		"&owner_id="+api.UserId+
		"&v="+api.Version, &response)
	return response.Response, err
}

func getAlbumNameById(id int, albums *[]Album) string {
	if albums == nil || id == 0 {
		return ""
	}

	for _, v := range *albums {
		if id == v.Id {
			return strings.Trim(v.Title, " ")
		}
	}
	return ""
}

func (api *VkApi) AudioGet(offset int, count int, albums *[]Album, base string) []Audio {
	var response ResponseGetAudio
	api.request("https://api.vk.com/method/audio.get"+
		"?access_token="+api.Token+
		"&owner_id="+api.UserId+
		"&album_id=0"+
		// "&audio_ids=1,2"
		"&need_user=0"+
		"&offset="+strconv.Itoa(offset)+
		"&count="+strconv.Itoa(count)+
		"&v="+api.Version, &response)
	for i, v := range response.Response.Items {
		v.Artist = html.UnescapeString(strings.Trim(v.Artist, " "))
		v.Title = html.UnescapeString(strings.Trim(v.Title, " "))
		v.Album = getAlbumNameById(v.AlbumId, albums)
		x := strings.Replace(v.Artist+" - "+v.Title+".mp3", "/", "-", -1)
		v.Path = filepath.Join(base, v.Album, x)
		response.Response.Items[i] = v
	}
	return response.Response.Items
}

func (api *VkApi) AudioGetAlbums() []Album {
	var response ResponseAlbums
	api.request("https://api.vk.com/method/audio.getAlbums"+
		"?owner_id="+api.UserId+
		"&count=100"+
		"&access_token="+api.Token+
		"&v="+api.Version, &response)
	return response.Response.Items
}

func (api *VkApi) AudioAddAlbum(title string) int {
	var response ResponseAudioAddAlbum
	api.request("https://api.vk.com/method/audio.addAlbum"+
		"?owner_id="+api.UserId+
		"&title="+title+
		"&access_token="+api.Token+
		"&v="+api.Version, &response)
	return response.Response.AlbumId
}

func (api *VkApi) AudioMoveToAlbum(album_id int, audio_ids string) bool {
	// hacky way to avert flood control
	time.Sleep(1 * time.Second)

	var response ResponseInt
	api.request("https://api.vk.com/method/audio.moveToAlbum"+
		"?owner_id="+api.UserId+
		"&album_id="+strconv.Itoa(album_id)+
		"&audio_ids="+audio_ids+
		"&access_token="+api.Token+
		"&v="+api.Version, &response)
	return response.Response == 1
}
