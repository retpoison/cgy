package cgy

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Channel struct {
	Name   string  `json:"name"`
	Videos []Video `json:"relatedStreams"`
}

type Video struct {
	Title            string    `json:"title"`
	Id               string    `json:"url"`
	Thumbnail        string    `json:"thumbnailUrl"`
	Uploader         string    `json:"uploaderName"`
	Uploaded         int       `json:"uploaded"`
	UploadDate       string    `json:"uploadedDate"`
	VidoeStreams     []Streams `json:"videoStreams"`
	AudioStreams     []Streams `json:"audioStreams"`
	Duration         int       `json:"duration"`
	FormatedDuration string
}

type Streams struct {
	Url       string `json:"url"`
	Quality   string `json:"quality"`
	Type      string `json:"mimeType"`
	VideoOnly bool   `json:"videoOnly"`
}

func getChannelVideos(instance, channelId string) Channel {
	var str, err = request(instance + "/channel/" + channelId)
	if err != nil {
		log.Println(err)
	}
	var channel = Channel{}
	getStruct(str, &channel)

	for i, v := range channel.Videos {
		channel.Videos[i].Id = strings.Split(v.Id, "v=")[1]
		channel.Videos[i].FormatedDuration = getDuration(v.Duration)
	}
	return channel
}

func getVideo(instance, videoId string) Video {
	var url string = fmt.Sprintf("%s/streams/%s",
		instance, url.QueryEscape(videoId))

	var str, err = request(url)
	if err != nil {
		log.Println(err)
	}
	var video = Video{}
	getStruct(str, &video)
	video.FormatedDuration = getDuration(video.Duration)

	return video
}

func getInstances() []string {
	var instances []string
	var url string = "https://raw.githubusercontent.com/wiki/TeamPiped/Piped-Frontend/Instances.md"
	var content, err = request(url)
	if err != nil {
		log.Println(err)
	}

	var skipped int = 0
	for _, v := range strings.Split(string(content), "\n") {
		var split = strings.Split(v, "|")
		if len(split) == 5 {
			if skipped < 2 {
				skipped++
				continue
			}
			instances = append(instances, strings.TrimSpace(split[1]))
		}
	}

	return instances
}

func requestInstances(ch chan []string, instance []string) {
	var wg sync.WaitGroup
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	var last = [][]string{}

	var req = func(in string) {
		defer wg.Done()
		var s []string
		start := time.Now()
		result, err := client.Get(in + "/streams/rRPQs_kM_nw")
		if err == nil {
			if result.StatusCode == 200 {
				s = []string{
					in,
					fmt.Sprintf("%.2fs", time.Since(start).Seconds()),
				}
				result.Body.Close()
				ch <- s
			}
		} else {
			s = []string{
				in,
				">5s",
			}
			last = append(last, s)
		}
	}

	for _, v := range instance {
		go req(v)
		wg.Add(1)
	}

	wg.Wait()
	for _, v := range last {
		ch <- v
	}
	close(ch)
}

func getDuration(duration int) string {
	return fmt.Sprintf("%02d:%02d:%02d",
		duration/60/60, duration/60%60, duration%60)
}

func request(url string) (string, error) {
	var resp, err = http.Get(url)
	if err != nil {
		return "", fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("request: %w", err)
	}
	var str = string(b)

	return str, nil
}

func getStruct(str string, v interface{}) error {
	var err = json.Unmarshal([]byte(str), &v)
	if err != nil {
		return fmt.Errorf("getStruct: %w", err)
	}
	return nil
}
