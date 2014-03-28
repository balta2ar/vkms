package main

import (
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	E_TOKEN_EXPIRED = 5
)

type VkApi struct {
	token   string
	userId  string
	version string
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
	Error *Error
}

type Error struct {
	ErrorCode int    `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
}

func getUrl(url string) ([]byte, error) {
	resp, err := http.Get(url)
	//fmt.Println(resp)
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
	//fmt.Println(body)
	ioutil.WriteFile("body.txt", body, 0600)
	return body, nil
}

func (api *VkApi) request(url string, obj interface{}) {
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
		log.Fatalln("Error is set, cannot continue")
	}
	err = json.Unmarshal(body, obj)
	if err != nil {
		log.Println(err)
		log.Fatalln("Cannot parse response as ConcreteType")
	}
}

func NewVkApi(appId string, permissions string, version string) *VkApi {
	authUrl := "https://oauth.vk.com/authorize" +
		"?client_id=" + appId +
		"&scope=" + permissions +
		"&redirect_uri=https://oauth.vk.com/blank.html" +
		"&display=page" +
		"&v=" + version +
		"&response_type=token"

	fmt.Println(authUrl)

	//token := "1cfaad4ed9fe6939fb"
	// accessUrl := "https://oauth.vk.com/blank.html#access_token=52ed549020d20ceb119697bb76acad7b250322f98d22026fbb237de039cad3155217b37eb524d964d05d3&expires_in=86400&user_id=1476677"
	//origAccessUrl := "https://oauth.vk.com/blank.html#access_token=fc1cdebd37d11fc960bbf5324da719b21944d233c03a23e61e2ccba9fc28ecc02f6c4105ba7ec9d349a54&expires_in=86400&user_id=1476677"
	//"https://oauth.vk.com/blank.html#access_token=197a1185e3ce26a57098c994de532f23bc2e8ab715356526f97c43d64817d523a1b69fd08a53c6fabb71a&expires_in=86400&user_id=1476677"
	origAccessUrl := "https://oauth.vk.com/blank.html#access_token=106c505b979f4738b1292735d58c8a106fb1e1e7058051d2323f934d9b98468113c928d5dc432cd16b2e7&expires_in=86400&user_id=1476677"
	accessUrl := strings.Replace(origAccessUrl, "#", "?", -1)
	parsedUrl, err := url.Parse(accessUrl)
	if err != nil {
		panic("Cannot parse accessUrl: " + accessUrl)
	}
	fmt.Println(parsedUrl)

	token := parsedUrl.Query().Get("access_token")
	userId := parsedUrl.Query().Get("user_id")
	fmt.Println("token:", token)
	fmt.Println("user_id:", userId)
	fmt.Println("expires_in:", parsedUrl.Query().Get("expires_in"))
	return &VkApi{token, userId, version}
}

func (api *VkApi) UsersGet(ids string) []User {
	url := "https://api.vk.com/method/users.get" +
		"?user_ids=" + ids +
		"&v=" + api.version

	body, err := getUrl(url)
	if err != nil {
		log.Fatal("Cannot get url")
	}

	var response ResponseUsersGet
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Println(err)
		log.Fatalln("Cannot get info about ids", ids)
	}
	return response.Response
}

type ResponseAudioGetCount struct {
	Response int
	Error    *Error
}

func (api *VkApi) AudioGetCount() int {
	var response ResponseAudioGetCount
	api.request("https://api.vk.com/method/audio.getCount"+
		"?access_token="+api.token+
		"&owner_id="+api.userId+
		"&v="+api.version, &response)
	return response.Response
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

func (api *VkApi) AudioGet(offset int, count int, albums *[]Album) []Audio {
	url := "https://api.vk.com/method/audio.get" +
		"?access_token=" + api.token +
		"&owner_id=" + api.userId +
		"&album_id=0" +
		// "&audio_ids=1,2"
		"&need_user=0" +
		"&offset=" + strconv.Itoa(offset) +
		"&count=" + strconv.Itoa(count) +
		"&v=" + api.version

	body, err := getUrl(url)
	if err != nil {
		log.Fatal("Cannot get audio")
	}

	var response ResponseGetAudio
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Println(err)
		log.Fatalln("Cannot parse audio get response")
	}

	for i, v := range response.Response.Items {
		v.Artist = html.UnescapeString(strings.Trim(v.Artist, " "))
		v.Title = html.UnescapeString(strings.Trim(v.Title, " "))
		v.Album = getAlbumNameById(v.AlbumId, albums)
		v.Path = strings.Replace(v.Artist+" - "+v.Title+".mp3", "/", "-", -1)
		response.Response.Items[i] = v
	}

	return response.Response.Items
}

type ResponseAlbums struct {
	Response struct {
		Count float64
		Items []Album
	}
}

func (api *VkApi) AudioGetAlbums() []Album {
	var response ResponseAlbums
	api.request("https://api.vk.com/method/audio.getAlbums"+
		"?owner_id="+api.userId+
		"&count=100"+
		"&access_token="+api.token+
		"&v="+api.version, &response)
	return response.Response.Items
}

func (api *VkApi) AudioAddAlbum(title string) int {
	url := "https://api.vk.com/method/audio.addAlbum" +
		"?owner_id=" + api.userId +
		"&title=" + title +
		"&access_token=" + api.token +
		"&v=" + api.version

	body, err := getUrl(url)
	if err != nil {
		log.Fatal("Cannot get url")
	}

	var response ResponseAudioAddAlbum
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Println(err)
		log.Fatalln("Cannot add album")
	}
	return response.Response.AlbumId
}

func (api *VkApi) AudioMoveToAlbum(album_id int, audio_ids string) bool {
	time.Sleep(1 * time.Second)

	url := "https://api.vk.com/method/audio.moveToAlbum" +
		"?owner_id=" + api.userId +
		"&album_id=" + strconv.Itoa(album_id) +
		"&audio_ids=" + audio_ids +
		"&access_token=" + api.token +
		"&v=" + api.version

	body, err := getUrl(url)
	if err != nil {
		log.Fatal("Cannot get url")
	}

	var response ResponseInt
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Println(err)
		log.Fatalln("Cannot move to ablum")
	}
	return response.Response == 1
}
