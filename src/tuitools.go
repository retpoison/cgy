package cgy

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/rivo/tview"
)

func refresh() {
	clearList(pagesMaps["channel"])
	clearList(pagesMaps["video"])
	addToList(pagesMaps["video"], "Refreshing...", "", nil)

	var videos []Video
	for _, ch := range config.Channels {
		addToList(pagesMaps["video"],
			fmt.Sprintf("Getting %s Videos...", ch), "", nil)
		app.Draw()

		channel, err := getChannelVideos(config.Instance, ch)
		if err != nil {
			log.Println(fmt.Errorf("refresh: %w", err))
			addToList(pagesMaps["channel"], "error:", err.Error(), nil)
			addToList(pagesMaps["video"], ch+" error:", err.Error(), nil)
			continue
		}

		videos = append(videos, channel.Videos...)
		addToList(pagesMaps["video"], "Done.", "", nil)
		addToList(pagesMaps["channel"], channel.Name, ch, nil)
	}

	clearList(pagesMaps["video"])
	sort.Slice(videos, func(i, j int) bool {
		return videos[i].Uploaded > videos[j].Uploaded
	})

	for _, v := range videos {
		addToList(pagesMaps["video"], v.Title,
			fmt.Sprintf("%s - %9s - %s - %s",
				pding(v.Uploader, 30), v.FormatedDuration, v.UploadDate, v.Id),
			nil)
	}

	app.Draw()
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
	log.Println("running:", command.String())
	err := command.Start()
	if err != nil {
		log.Println("error running player:", err)
	}
}

func getArgs(title, url, audioUrl, thumbnail string) []string {
	var args = []string{}
	if audioUrl == "" {
		for _, v := range config.Options {
			if v == "--audio-file=%audio%" {
				continue
			}
			args = append(args, v)
		}
	} else {
		args = config.Options
	}

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
	list.(*tview.List).SetCurrentItem(0)
	list.(*tview.List).Clear()
}

func addChannel(channel string) {
	var chs []string
	chs = append(config.Channels, channel)
	err := config.set("Channels", chs)
	if err != nil {
		log.Println(err)
	}
}

func removeChannel(channel string) {
	var chs []string
	for _, v := range config.Channels {
		if v != channel {
			chs = append(chs, v)
		}
	}
	err := config.set("Channels", chs)
	if err != nil {
		log.Println(err)
	}
}
