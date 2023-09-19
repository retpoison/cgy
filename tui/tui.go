package tui

import (
	"cgy/config"
	"fmt"
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	app       *tview.Application
	pages     *tview.Pages
	configs   *config.Config
	pagesMaps map[string]tview.Primitive
)

func init() {
	pages = getPage()
	app = tview.NewApplication()
	app.SetRoot(pages, true)

	var refresh = func() {
		refreshVideos()
		refreshChannels()
	}

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		pageName, _ := pages.GetFrontPage()
		if pageName == "addChannel" {
			return event
		}

		switch event.Rune() {
		case 'V', 'v':
			pages.SwitchToPage("video")
		case 'C', 'c':
			pages.SwitchToPage("channel")
		case 'A', 'a':
			pages.SwitchToPage("addChannel")
			return nil
		case 'R', 'r':
			go refresh()
		case 'I', 'i':
			pages.SwitchToPage("instance")
			go updateInstances()
		case 'H', 'h':
			pages.SwitchToPage("help")
		case 'Q', 'q':
			app.Stop()
		}

		return event
	})

	go refresh()
}

func Run(conf *config.Config) {
	configs = conf

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

func getPage() *tview.Pages {
	var videoList = tview.NewList().
		SetSelectedFunc(selectedVideo)
	videoList.SetBorder(true).SetTitle(" Videos ═══ press h for help ")

	var selectedChannel string
	var channelList = tview.NewList().
		ShowSecondaryText(false).
		SetSelectedFunc(func(_ int, mainText string, _ string, _ rune) {
			selectedChannel = mainText
			pagesMaps["delete"].(*tview.Modal).SetText(
				fmt.Sprintf("Do you want to remove\n%s ?", selectedChannel))
			pages.SwitchToPage("delete")
		})
	channelList.SetBorder(true).SetTitle(" Channels ═══ press h for help ")

	var helpText = tview.NewTextView().
		SetRegions(true).
		SetText(getHelpText()).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEscape {
				pages.SwitchToPage("video")
			}
		})
	helpText.SetBorder(true)

	var channelInput = tview.NewInputField().
		SetLabel("Enter channel Id: ").
		SetFieldWidth(30)
	channelInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			configs.AddChannel(channelInput.GetText())
		}
		channelInput.SetText("")
		pages.SwitchToPage("video")
	})
	channelInput.SetBorder(true).
		SetTitle(" Add channel ")

	var deleteButton = tview.NewModal().
		AddButtons([]string{"No", "Yes"}).
		SetDoneFunc(func(_ int, buttonLabel string) {
			if buttonLabel == "Yes" {
				configs.RemoveChannel(getId(selectedChannel))
				go refreshChannels()
				pages.SwitchToPage("channel")
			} else if buttonLabel == "No" {
				pages.SwitchToPage("channel")
			}
		})

	var qualityList = tview.NewList().
		ShowSecondaryText(false).
		SetDoneFunc(func() {
			pages.SwitchToPage("video")
		})
	qualityList.SetBorder(true).SetTitle(" Choose quality ")

	var instanceList = tview.NewList().
		ShowSecondaryText(false).
		SetSelectedFunc(func(_ int, mainText string, _ string, _ rune) {
			configs.Set("instance", mainText)
			pages.SwitchToPage("video")
		}).
		SetDoneFunc(func() {
			pages.SwitchToPage("video")
		})
	instanceList.SetBorder(true).SetTitle(" Instances ")

	pagesMaps = map[string]tview.Primitive{}
	pagesMaps["video"] = videoList
	pagesMaps["channel"] = channelList
	pagesMaps["help"] = helpText
	pagesMaps["addChannel"] = channelInput
	pagesMaps["delete"] = deleteButton
	pagesMaps["quality"] = qualityList
	pagesMaps["instance"] = instanceList

	var pages = tview.NewPages().
		AddPage("video", pagesMaps["video"], true, true).
		AddPage("channel", pagesMaps["channel"], true, false).
		AddPage("help", center(pagesMaps["help"], 30, 20), true, false).
		AddPage("addChannel", center(pagesMaps["addChannel"], 60, 3), true, false).
		AddPage("delete", center(pagesMaps["delete"], 60, 3), true, false).
		AddPage("quality", pagesMaps["quality"], true, false).
		AddPage("instance", pagesMaps["instance"], true, false)

	return pages
}
