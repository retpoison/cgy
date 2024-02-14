package cgy

import (
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	app       *tview.Application
	pages     *tview.Pages
	pagesMaps map[string]tview.Primitive
)

func InitTui() {
	pages = getPage()
	app = tview.NewApplication()
	app.SetRoot(pages, true)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		pageName, _ := pages.GetFrontPage()
		if pageName == "addChannel" || pageName == "play" {
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
		case 'P', 'p':
			pages.SwitchToPage("play")
			return nil
		case 'R', 'r':
			go refreshVideos()
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
}

func run() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
