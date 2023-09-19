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

	for _, ch := range configs.Channels {
		var channel = piped.GetChannelVideos(configs.Instance, ch)
		for _, v := range channel.Videos {
			addToList(pagesMaps["video"], v.Title,
				fmt.Sprintf("%s - %9s - %s - %s",
					pding(channel.Name, 30), v.FormatedDuration, v.UploadDate, v.Id),
				nil)
		}
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

func selectedVideo(_ int, mainText string, secondaryText string, _ rune) {
	clearList(pagesMaps["quality"])
	pagesMaps["quality"].(*tview.List).SetTitle(mainText)
	pages.SwitchToPage("quality")

	addToList(pagesMaps["quality"], "Getting available qualities...", "", nil)

	go func() {
		var split = strings.Split(secondaryText, " ")
		var video = piped.GetVideo(configs.Instance, split[len(split)-1])
		clearList(pagesMaps["quality"])
		pagesMaps["quality"].(*tview.List).SetSelectedFunc(func(index int, _ string, _ string, _ rune) {
			playStream(video.Title, findUrl(index, video))
		})
		var txt string
		for _, v := range video.VidoeStreams {
			txt = fmt.Sprintf("%-12s %-10s", v.Type, v.Quality)
			if v.VideoOnly == true {
				txt += " video Only"
			}
			addToList(pagesMaps["quality"], txt, "", nil)
		}
		for _, v := range video.AudioStreams {
			txt = fmt.Sprintf("%-12s %-10s", v.Type, v.Quality)
			if v.VideoOnly == true {
				txt += " video Only"
			}
			addToList(pagesMaps["quality"], txt, "", nil)
		}
		app.Draw()
	}()
}

func playStream(title, url string) {
	var args = strings.Split(configs.Options, " ")
	for i, v := range args {
		args[i] = strings.Replace(v, "%title%", title, 1)
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
S, s    Setting
H, h    help
Q, q    Quit

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
