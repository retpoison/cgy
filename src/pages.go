package cgy

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func getPage() *tview.Pages {
	var selectedChannel, selectedChannelID string
	pagesMaps = map[string]tview.Primitive{}
	pagesMaps["video"] = getVideoList()
	pagesMaps["channel"] = getChannelList(&selectedChannel, &selectedChannelID)
	pagesMaps["help"] = getHelpBox()
	pagesMaps["addChannel"] = getChannelInput()
	pagesMaps["play"] = getPlayInput()
	pagesMaps["delete"] = getDeleteChannel(&selectedChannel, &selectedChannelID)
	pagesMaps["quality"] = getQualityList()
	pagesMaps["instance"] = getInstanceList()

	var pages = tview.NewPages().
		AddPage("video", pagesMaps["video"], true, true).
		AddPage("channel", pagesMaps["channel"], true, false).
		AddPage("help", center(pagesMaps["help"], helpTextWidth(), helpTextHeight()), true, false).
		AddPage("addChannel", center(pagesMaps["addChannel"], 60, 3), true, false).
		AddPage("play", center(pagesMaps["play"], 85, 3), true, false).
		AddPage("delete", center(pagesMaps["delete"], 60, 3), true, false).
		AddPage("quality", pagesMaps["quality"], true, false).
		AddPage("instance", pagesMaps["instance"], true, false)

	return pages
}

func getVideoList() tview.Primitive {
	var videoList = tview.NewList()
	videoList.AddItem("press R,r to refresh.", "", 0, nil).
		SetBorder(true).
		SetTitle(" Videos ══ h for help ").
		SetInputCapture(vimShortcuts)

	videoList.SetSelectedFunc(func(index int, _, secondaryText string, _ rune) {
		if secondaryText == "" {
			return
		}
		var id, err = getVideoId(index, secondaryText)
		if err != nil {
			pages.SwitchToPage("video")
		} else {
			qualities(id)
		}
	})
	return videoList
}

func getChannelList(sChannel, sChannelID *string) tview.Primitive {
	var channelList = tview.NewList()
	channelList.AddItem("press R,r to refresh.", "", 0, nil).
		SetBorder(true).
		SetTitle(" Channels ══ h for help ").
		SetInputCapture(vimShortcuts)

	channelList.SetSelectedFunc(func(_ int, mainText string, secondaryText string, _ rune) {
		*sChannel = mainText
		*sChannelID = secondaryText
		pagesMaps["delete"].(*tview.Modal).SetText(
			fmt.Sprintf("Do you want to remove %s?", *sChannel))
		pages.SwitchToPage("delete")
	})
	return channelList
}

func getHelpBox() tview.Primitive {
	var helpBox = tview.NewTextView()
	helpBox.SetRegions(true).
		SetText(getHelpText()).
		SetBorder(true)

	helpBox.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			pages.SwitchToPage("video")
		}
	})
	return helpBox
}

func getChannelInput() tview.Primitive {
	var channelInput = tview.NewInputField()
	channelInput.SetLabel("Enter channel Id: ").
		SetFieldWidth(30).
		SetBorder(true).
		SetTitle(" Add channel ")

	channelInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			config.addChannel(channelInput.GetText())
		}
		channelInput.SetText("")
		go refresh()
		pages.SwitchToPage("video")
	})
	return channelInput
}

func getPlayInput() tview.Primitive {
	var playInput = tview.NewInputField()
	playInput.SetLabel("Enter video Id or url: ").
		SetFieldWidth(60).
		SetBorder(true).
		SetTitle(" Play ")

	playInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			var id, err = getVideoId(-1, playInput.GetText())
			if err != nil {
				pages.SwitchToPage("video")
			} else {
				qualities(id)
			}
		} else {
			pages.SwitchToPage("video")
		}
		playInput.SetText("")
	})
	return playInput
}

func getDeleteChannel(sChannel, sChannelID *string) tview.Primitive {
	var deleteChannel = tview.NewModal()
	deleteChannel.AddButtons([]string{"No", "Yes"})

	deleteChannel.SetDoneFunc(func(_ int, buttonLabel string) {
		if buttonLabel == "Yes" {
			config.removeChannel(*sChannelID)
			go refresh()
			pages.SwitchToPage("channel")
		} else if buttonLabel == "No" {
			pages.SwitchToPage("channel")
		}
	})
	return deleteChannel
}

func getQualityList() tview.Primitive {
	var qualityList = tview.NewList()
	qualityList.ShowSecondaryText(false).
		SetDoneFunc(func() { pages.SwitchToPage("video") }).
		SetBorder(true).
		SetTitle(" Choose quality ").
		SetInputCapture(vimShortcuts)
	return qualityList
}

func getInstanceList() tview.Primitive {
	var instanceList = tview.NewList()
	instanceList.SetDoneFunc(func() { pages.SwitchToPage("video") }).
		SetBorder(true).
		SetTitle(" Instances ").
		SetInputCapture(vimShortcuts)

	instanceList.SetSelectedFunc(func(_ int, mainText string, _ string, _ rune) {
		config.set("Instance", mainText)
		pages.SwitchToPage("video")
	})
	return instanceList
}

func getHelpText() string {
	return `Keyboard shortcuts:
  j  |-------------Down
  k  |---------------Up
  g  |Beginning of list
  G  |------End of list
V, v |-----------Videos
C, c |---------Channels
A, a |------Add channel
P, p |-Play given video
R, r |----------Refresh
I, i |--------Instances
H, h |-------------help
Q, q |-------------Quit
Esc  |-------------Back

Enter, Space:
Play in video list
Delete in channel list
Select in instance list
Select in quality list`
}

func vimShortcuts(e *tcell.EventKey) *tcell.EventKey {
	switch e.Rune() {
	case 'j':
		return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
	case 'k':
		return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
	case 'g':
		return tcell.NewEventKey(tcell.KeyHome, 0, tcell.ModNone)
	case 'G':
		return tcell.NewEventKey(tcell.KeyEnd, 0, tcell.ModNone)
	}
	return e
}
