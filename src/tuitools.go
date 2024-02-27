package cgy

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"unicode/utf8"

	"github.com/rivo/tview"
)

func refreshVideos() {
	clearList(pagesMaps["video"])
	addToList(pagesMaps["video"], "Refreshing...", "", nil)
	channels := refreshChannels()

	var videosSlice [][]Video
	var videosCount int = 0
	var sortedChan = make(chan Video, 3)

	for chName, chID := range channels {
		addToList(pagesMaps["video"],
			fmt.Sprintf("Getting %s Videos...", chName), "", nil)
		app.Draw()

		v, err := getChannelVideos(config.Instance, chID)
		if err != nil {
			log.Println(fmt.Errorf("refreshVideos: %w", err))
			addToList(pagesMaps["video"], chName+" error:", err.Error(), nil)
			app.Draw()
			continue
		}
		videosSlice = append(videosSlice, v.Videos)
		videosCount += len(videosSlice[len(videosSlice)-1])

		addToList(pagesMaps["video"], "Done.", "", nil)
	}

	go sortVideos(sortedChan, videosSlice, videosCount)
	clearList(pagesMaps["video"])
	for v := range sortedChan {
		addToList(pagesMaps["video"], v.Title,
			fmt.Sprintf("%s - %9s - %s - %s",
				pding(v.Uploader, 30), v.FormatedDuration, v.UploadDate, v.Id),
			nil)
	}

	app.Draw()
}

func refreshChannels() map[string]string {
	clearList(pagesMaps["channel"])

	var channels = map[string]string{}
	for _, ch := range config.Channels {
		channel, err := getChannelVideos(config.Instance, ch)
		if err != nil {
			log.Println(fmt.Errorf("refreshChannels: %w", err))
			addToList(pagesMaps["channel"], "error:", err.Error(), nil)
			app.Draw()
			continue
		}

		channels[channel.Name] = ch

		addToList(pagesMaps["channel"], channel.Name, ch, nil)
	}
	app.Draw()

	return channels
}

func updateInstances() {
	pagesMaps["instance"].(*tview.List).
		SetTitle(" Instances ═══ " + config.Instance + " ")
	clearList(pagesMaps["instance"])
	addToList(pagesMaps["instance"], "Getting instances...", "", nil)
	instances, err := getInstances()
	clearList(pagesMaps["instance"])
	if err != nil {
		log.Println(fmt.Errorf("updateInstances: %w", err))
		addToList(pagesMaps["instance"], "error: ", err.Error(), nil)
		app.Draw()
		return
	}

	var ch = make(chan []string, 2)
	go requestInstances(ch, instances)
	for v := range ch {
		addToList(pagesMaps["instance"], v[0], v[1], nil)
		app.Draw()
	}
}

func sortVideos(schan chan<- Video, videos [][]Video, count int) {
	defer close(schan)
	var index = make([]int, len(videos))
	for i := range index {
		index[i] = 0
	}

	for i := 0; i < count; i++ {
		max := 0
		for j := 0; j < len(videos); j++ {
			if index[j] >= len(videos[j]) {
				continue
			}
			if videos[j][index[j]].Uploaded > max {
				max = videos[j][index[j]].Uploaded
			}
		}
		for j := 0; j < len(videos); j++ {
			if index[j] >= len(videos[j]) {
				continue
			}
			if videos[j][index[j]].Uploaded == max {
				schan <- videos[j][index[j]]
				index[j]++
			}
		}
	}
}

func qualities(id string) {
	pagesMaps["quality"].(*tview.List).SetTitle(" Choose quality ")
	clearList(pagesMaps["quality"])
	pages.SwitchToPage("quality")
	addToList(pagesMaps["quality"], "Getting available qualities...", "", nil)

	go func() {
		video, err := getVideo(config.Instance, id)
		if err != nil {
			log.Println(fmt.Errorf("qualities: %w", err))
			addToList(pagesMaps["quality"], "Error:"+err.Error(), "", nil)
			app.Draw()
			return
		}

		clearList(pagesMaps["quality"])
		pagesMaps["quality"].(*tview.List).SetTitle(" " + video.Title + " ")
		pagesMaps["quality"].(*tview.List).SetSelectedFunc(func(index int, mainText string, _ string, _ rune) {
			if strings.Contains(mainText, "video only") {
				playStream(getArgs(video.Title, findUrl(index, video),
					video.AudioStreams[0].Url, video.Thumbnail))
			} else {
				playStream(getArgs(video.Title, findUrl(index, video), "", video.Thumbnail))
			}
		})
		var txt string
		for _, v := range video.VidoeStreams {
			txt = fmt.Sprintf("%-12s %-10s", v.Type, v.Quality)
			if v.VideoOnly {
				txt += " video only"
			}
			addToList(pagesMaps["quality"], txt, "", nil)
		}
		for _, v := range video.AudioStreams {
			txt = fmt.Sprintf("%-12s %-10s", v.Type, v.Quality)
			addToList(pagesMaps["quality"], txt, "", nil)
		}
		app.Draw()
	}()
}

func playStream(args []string) {
	var command = exec.Command(config.Program, args...)
	command.Start()
}

func getArgs(title, url, audioUrl, thumbnail string) []string {
	var options string
	if audioUrl == "" {
		options = strings.Replace(config.Options,
			"--audio-file=%audio%", "", 1)
	} else {
		options = config.Options
	}

	var args = strings.Split(options, " ")
	for i, v := range args {
		if strings.Contains(v, "%title%") {
			args[i] = strings.Replace(v, "%title%", title, 1)
		} else if strings.Contains(v, "%audio%") {
			args[i] = strings.Replace(v, "%audio%", audioUrl, 1)
		} else if strings.Contains(v, "%thumbnail%") {
			args[i] = strings.Replace(v, "%thumbnail%", thumbnail, 1)
		}
	}

	args = append(args, url)
	return args
}

func findUrl(index int, v Video) string {
	if index <= len(v.VidoeStreams)-1 {
		return v.VidoeStreams[index].Url
	}
	return v.AudioStreams[index-len(v.VidoeStreams)].Url
}

func pding(str string, length int) string {
	if utf8.RuneCountInString(str) > length {
		var rstr = []rune(str)[:length-3]
		str = string(rstr) + "..."
	} else {
		str = str + strings.Repeat(" ", length-utf8.RuneCountInString(str))
	}
	return str
}

func helpTextWidth() int {
	var max int = 0
	for _, v := range strings.Split(getHelpText(), "\n") {
		if len(v) > max {
			max = len(v)
		}
	}
	return max + 3
}

func helpTextHeight() int {
	return len(strings.Split(getHelpText(), "\n")) + 3
}

func center(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(p, height, 1, true).
			AddItem(nil, 0, 1, false), width, 1, true).
		AddItem(nil, 0, 1, false)
}

func getVideoId(status int, str string) (string, error) {
	if status != -1 {
		var split = strings.Split(str, " ")
		return split[len(split)-1], nil
	}

	// YouTube ID is a string of 11 characters.
	var id string = ""
	if len(str) == 11 {
		id = str
	} else if strings.Contains(str, "v=") {
		var index = strings.Index(str, "v=")
		id = str[index+2 : index+13]
	} else if len(str) > 11 {
		var split = strings.Split(str, "/")
		id = split[len(split)-1]
	}

	if strings.Contains(id, "&") ||
		strings.Contains(id, "/") ||
		len(id) < 11 {
		return "", errors.New("Short or wrong video id")
	}
	return id, nil
}

func addToList(list tview.Primitive, mainText, secondaryText string, selected func()) {
	list.(*tview.List).AddItem(mainText, secondaryText, 0, selected)
}

func clearList(list tview.Primitive) {
	list.(*tview.List).Clear()
}
