package tui

import (
	"fmt"
	"os/exec"
	"strings"
	"unicode/utf8"

	"cgy/piped"

	"github.com/rivo/tview"
)

func addToList(list tview.Primitive, mainText, secondaryText string, selected func()) {
	list.(*tview.List).AddItem(mainText, secondaryText, 0, selected)
}

func clearList(list tview.Primitive) {
	list.(*tview.List).Clear()
}

func refreshVideos() {
	clearList(pagesMaps["video"])
	addToList(pagesMaps["video"], "Getting Videos...", "", nil)
	var videosSlice [][]piped.Video
	var videosCount int = 0
	var sortedChan = make(chan piped.Video, 3)

	for i, ch := range configs.Channels {
		videosSlice = append(videosSlice, piped.GetChannelVideos(configs.Instance, ch).Videos)
		videosCount += len(videosSlice[i])
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

func refreshChannels() {
	clearList(pagesMaps["channel"])

	for _, ch := range configs.Channels {
		var channel = piped.GetChannelVideos(configs.Instance, ch)

		addToList(pagesMaps["channel"],
			fmt.Sprintf("%-20s %s", channel.Name, ch), "", nil)
	}
	app.Draw()
}

func updateInstances() {
	clearList(pagesMaps["instance"])
	addToList(pagesMaps["instance"], "Getting instances...", "", nil)
	var instances = piped.GetInstances()
	clearList(pagesMaps["instance"])
	for _, v := range instances {
		addToList(pagesMaps["instance"], v, "", nil)
	}
	app.Draw()
}

func sortVideos(schan chan<- piped.Video, videos [][]piped.Video, count int) {
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

func selectedVideo(index int, _, secondaryText string, _ rune) {
	var id, err = getVideoId(index, secondaryText)
	if err != nil {
		pages.SwitchToPage("video")
	} else {
		qualities(id)
	}
}

func qualities(id string) {
	clearList(pagesMaps["quality"])
	pages.SwitchToPage("quality")

	addToList(pagesMaps["quality"], "Getting available qualities...", "", nil)

	go func() {
		var video = piped.GetVideo(configs.Instance, id)
		clearList(pagesMaps["quality"])
		pagesMaps["quality"].(*tview.List).SetTitle(" " + video.Title + " ")
		pagesMaps["quality"].(*tview.List).SetSelectedFunc(func(index int, mainText string, _ string, _ rune) {
			if strings.Contains(mainText, "video only") {
				playStream(getArgs(video.Title, findUrl(index, video),
					video.AudioStreams[0].Url))
			} else {
				playStream(getArgs(video.Title, findUrl(index, video), ""))
			}
		})
		var txt string
		for _, v := range video.VidoeStreams {
			txt = fmt.Sprintf("%-12s %-10s", v.Type, v.Quality)
			if v.VideoOnly == true {
				txt += " video only"
			}
			addToList(pagesMaps["quality"], txt, "", nil)
		}
		for _, v := range video.AudioStreams {
			txt = fmt.Sprintf("%-12s %-10s", v.Type, v.Quality)
			if v.VideoOnly == true {
				txt += " video only"
			}
			addToList(pagesMaps["quality"], txt, "", nil)
		}
		app.Draw()
	}()
}

func playStream(args []string) {
	var command = exec.Command(configs.Program, args...)
	command.Start()
}

func getArgs(title, url, audioUrl string) []string {
	var options string
	if audioUrl == "" {
		options = strings.Replace(configs.Options,
			"--audio-file=%audio%", "", 1)
	} else {
		options = configs.Options
	}

	var args = strings.Split(options, " ")
	for i, v := range args {
		if strings.Contains(v, "%title%") {
			args[i] = strings.Replace(v, "%title%", title, 1)
		} else if strings.Contains(v, "%audio%") {
			args[i] = strings.Replace(v, "%audio%", audioUrl, 1)
		}
	}

	args = append(args, url)
	return args
}

func findUrl(index int, v piped.Video) string {
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

func getHelpText() string {
	return `Keyboard shortcuts:
V, v    Videos
C, c    Channels
A, a    Add channel
P, p    Play given video
R, r    Refresh videos
I, i    Instances
H, h    help
Q, q    Quit
Esc     Back

Press Enter on video for playing
Press Enter on channel for deleting`
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

func getChannelId(str string) string {
	var split = strings.Split(str, " ")
	return split[len(split)-1]
}

func getVideoId(status int, str string) (id string, err error) {
	id = ""
	err = nil
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	if status != -1 {
		var split = strings.Split(str, " ")
		id = split[len(split)-1]
		return
	}

	// YouTube ID is a string of 11 characters.
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
		id = ""
		err = fmt.Errorf("Short or wrong video id")
	}
	return
}
