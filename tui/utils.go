package tui

import (
	"cgy/piped"
	"fmt"
	"os/exec"
	"strings"
	"unicode/utf8"

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

func selectedVideo(_ int, mainText string, secondaryText string, _ rune) {
	clearList(pagesMaps["quality"])
	pagesMaps["quality"].(*tview.List).SetTitle(mainText)
	pages.SwitchToPage("quality")

	addToList(pagesMaps["quality"], "Getting available qualities...", "", nil)

	go func() {
		var split = strings.Split(secondaryText, " ")
		var video = piped.GetVideo(configs.Instance, split[len(split)-1])
		clearList(pagesMaps["quality"])
		pagesMaps["quality"].(*tview.List).SetSelectedFunc(func(index int, mainText string, _ string, _ rune) {
			if strings.Contains(mainText, "video only") {
				playStream(video.Title, findUrl(index, video),
					video.AudioStreams[0].Url)
			} else {
				playStream(video.Title, findUrl(index, video), "")
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

func playStream(title, url, audioUrl string) {
	var options string
	if audioUrl == "" {
		options = strings.Replace(configs.Options,
			"--audio-file=%audio%", "", 1)
	} else {
		options = configs.Options
	}

	var args = strings.Split(options, " ")
	for i, v := range args {
		args[i] = strings.Replace(v, "%title%", title, 1)
		args[i] = strings.Replace(v, "%audio%", audioUrl, 1)
	}

	args = append(args, url)

	var command = exec.Command(configs.Program, args...)
	command.Start()
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
R, r    Refresh videos
I, i    Instances
H, h    help
Q, q    Quit
Esc     Back

Press Enter on video for playing
Press Enter on channel fo deleting`
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

func getId(str string) string {
	var split = strings.Split(str, " ")
	return split[len(split)-1]
}
