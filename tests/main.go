package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

const (
	baseURL    = "http://localhost:8082"
	sm9TestTag = "今年レンコンコマンダー常盤"
)

func main() {
	// Can we login?
	client := authenticate("admin", "admin")
	log.Println("Logged in successfully")
	// Can we try to archive something?
	// bilibili is currently broken...
	//makeArchiveRequest(client, "bilibili", "tag", "sm35952346")
	//makeArchiveRequest(client, "bilibili", "channel", "1963331522")

	makeArchiveRequest(client, "niconico", "tag", "TEST_sm9")
	makeArchiveRequest(client, "niconico", "channel", "119163275")
	makeArchiveRequest(client, "niconico", "playlist", "58583228")

	makeArchiveRequest(client, "youtube", "channel", "UCF43Xa8ZNQqKs1jrhxlntlw")            // Some random channel I found with short videos. Good enough!
	makeArchiveRequest(client, "youtube", "playlist", "PLz2PzeiUFQLuZ6k_e50OEK0xd_NAy7xat") // random playlist with one entry

	// Are videos being downloaded and transcoded correctly?
	for start := time.Now(); time.Since(start) < time.Minute*30; {
		time.Sleep(time.Second * 30)
		//err := pageHasVideos(client, "sm35952346", 1) // Bilibili tag
		//if err != nil {
		//	log.Println(err)
		//	continue
		//}
		//
		//err = pageHasVideos(client, "被劝诱的石川", 1) // Bilibili channel
		//if err != nil {
		//	log.Println(err)
		//	continue
		//}

		err := pageHasVideos(client, "風野灯織", 1) // nico channel
		if err != nil {
			log.Println(err)
			continue
		}

		err = pageHasVideos(client, "中の", 1) // there's some bizarre nico bug here where the tags keep switching on the video. very strange
		if err != nil {
			log.Println(err)
			continue
		}

		err = pageHasVideos(client, "しゅんなな", 8) // yt channel, should be 13 but several have ffmpeg errors. Sad!
		if err != nil {
			log.Println(err)
			continue
		}

		err = pageHasVideos(client, "untitled_0360", 1) // yt playlist
		if err != nil {
			log.Println(err)
			continue
		}

		err = pageHasVideos(client, "NEW+GAME", 1) // Nico mylist
		if err != nil {
			log.Println(err)
			continue
		}

		log.Println("All videos downloaded and transcoded successfully")
		return
	}

	log.Panic("Failed to download and transocde videos within 30 minutes")

}

func pageHasVideos(client *http.Client, tag string, count int) error {
	response, _ := client.Get(baseURL + fmt.Sprintf("/?search=%s&category=upload_date", tag))
	cont, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	c := strings.Count(string(cont), "href=\"/videos/")
	if c < count {
		return fmt.Errorf("page does not contain the right number of videos for %s. Found: %d", tag, c)
	}

	return nil
}

func makeArchiveRequest(client *http.Client, website, contentType, contentValue string) {
	response, _ := client.PostForm(baseURL+"/archiverequests", url.Values{
		"website":      {website},
		"contentType":  {contentType},
		"contentValue": {contentValue},
	})

	if response.StatusCode != 301 {
		log.Fatalf("bad archival request response status: %d", response.StatusCode)
	}

	return
}

var redirectErr error = errors.New("don't redirect")

func authenticate(username, password string) *http.Client {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return redirectErr
		},
		Jar: jar,
	}

	response, _ := client.PostForm(baseURL+"/login", url.Values{
		"username": {username},
		"password": {password},
	})
	// lol how do i check for an error here? :thinking:
	// if err != nil && err != redirectErr {
	// 	log.Panicf("failed to post with err: %s", err)
	// }

	if response.StatusCode != 301 {
		log.Panicf("bad auth status code: %d", response.StatusCode)
	}

	jwt := ""
	for _, cookie := range response.Cookies() {
		if cookie.Name == "jwt" {
			jwt = cookie.Value
		}
	}

	if jwt == "" {
		log.Panicf("jwt cookie not set")
	}

	return client
}
