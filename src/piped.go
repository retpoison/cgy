package cgy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Channel struct {
	Name    string  `json:"name"`
	Videos  []Video `json:"relatedStreams"`
	Message string  `json:"message"`
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
	Message          string    `json:"message"`
	FormatedDuration string
}

type Streams struct {
	Url       string `json:"url"`
	Quality   string `json:"quality"`
	Type      string `json:"mimeType"`
	VideoOnly bool   `json:"videoOnly"`
}

type Instance struct {
	Url string `json:"api_url"`
}

func getChannelVideos(instance, channelId string) (Channel, error) {
	var channel = Channel{}
	var str, err = request(instance + "/channel/" + channelId)
	if err != nil {
		return channel, fmt.Errorf("getChannelVideos: %w", err)
	}
	err = getStruct(str, &channel)
	if err != nil {
		return channel, fmt.Errorf("getChannelVideos: %w", err)
	} else if strings.Contains(channel.Message, "This channel does not exist.") {
		return channel, fmt.Errorf("getChannelVideos: channel with id:%s does not exist.", channelId)
	}

	for i, v := range channel.Videos {
		channel.Videos[i].Id = strings.Split(v.Id, "v=")[1]
		channel.Videos[i].FormatedDuration = getDuration(v.Duration)
	}
	return channel, nil
}

func getVideo(instance, videoId string) (Video, error) {
	var video = Video{}
	var url string = fmt.Sprintf("%s/streams/%s",
		instance, url.QueryEscape(videoId))
	var str, err = request(url)
	if err != nil {
		return video, fmt.Errorf("getVideo: %w", err)
	}
	err = getStruct(str, &video)
	if err != nil {
		return video, fmt.Errorf("getVideo: %w", err)
	} else if video.Message == "Video unavailable" {
		return video, fmt.Errorf("getVideo: video with id:%s is unavailable", videoId)
	}
	video.FormatedDuration = getDuration(video.Duration)

	return video, nil
}

func getInstances() ([]string, error) {
	var in bool = false
	var instances = []string{}
	var url string = "https://raw.githubusercontent.com/TeamPiped/documentation/main/content/docs/public-instances/index.md"
	var content, err = request(url)
	if err != nil {
		goto second_way
	}

	for _, v := range strings.Split(string(content), "\n") {
		var split = strings.Split(v, "|")
		if len(split) == 5 {
			if strings.TrimSpace(split[0]) == "---" {
				in = true
				continue
			}
			if in {
				instances = append(instances, strings.TrimSpace(split[1]))
			}
		}
	}
	return instances, nil

second_way:
	var instances_struct = []Instance{}
	url = "https://piped-instances.kavin.rocks/"
	content, err = request(url)
	if err != nil {
		return instances, fmt.Errorf("getInstances: %w", err)
	}

	err = getStruct(content, &instances_struct)
	if err != nil {
		return instances, fmt.Errorf("getInstances: %w", err)
	}
	for _, v := range instances_struct {
		instances = append(instances, v.Url)
	}

	return instances, nil
}

func requestInstances(ch chan []string, instance []string) {
	var wg sync.WaitGroup
	var mx sync.RWMutex
	var transport *http.Transport
	var proxy *url.URL

	if config.Proxy == "" {
		transport = http.DefaultTransport.(*http.Transport)
	} else {
		proxy, _ = url.Parse(config.Proxy)
		transport = &http.Transport{Proxy: http.ProxyURL(proxy)}
	}

	client := http.Client{
		Timeout:   5 * time.Second,
		Transport: transport,
	}
	var last = [][]string{}

	var req = func(in string) {
		defer wg.Done()
		var s []string
		request, _ := http.NewRequest("GET",
			in+"/streams/rRPQs_kM_nw", nil)
		request.Header.Set("User-Agent",
			"Mozilla/5.0 (Windows NT 10.0; rv:112.0) Gecko/20100101 Firefox/112.0  uacq")

		start := time.Now()
		result, err := client.Do(request)
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
				fmt.Sprintf(">5s code: %s", err.Error()),
			}
			mx.Lock()
			last = append(last, s)
			mx.Unlock()
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

func request(req_url string) (string, error) {
	var transport *http.Transport
	var proxy *url.URL
	var err error

	if config.Proxy == "" {
		transport = http.DefaultTransport.(*http.Transport)
	} else {
		proxy, err = url.Parse(config.Proxy)
		if err != nil {
			return "", fmt.Errorf("request: %w", err)
		}
		transport = &http.Transport{Proxy: http.ProxyURL(proxy)}
	}

	var client = http.Client{Transport: transport}
	req, err := http.NewRequest("GET", req_url, nil)
	if err != nil {
		return "", fmt.Errorf("request: %w", err)
	}
	req.Header.Set("User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; rv:112.0) Gecko/20100101 Firefox/112.0  uacq")

	resp, err := client.Do(req)
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
