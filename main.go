package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"strings"
	"time"
)

// helper function to create a popup like window mid screen
func popup(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(p, height, 1, false).
			AddItem(nil, 0, 1, false), width, 1, false).
		AddItem(nil, 0, 1, false)
}

func a(app *tview.Application) {
	for {
		time.Sleep(10)
		app.Draw()
	}
}

func main() {

	localUser := user{active: true}
	loggedUsersUpdate := make(chan bool, 10)
	conns := make(connections)
	conns["local"] = &localUser

	app := tview.NewApplication()

	go a(app) // WHY IS THIS NEEDED ?!?!?!?!?!?!?!?!?!?!

	// window showing logged users to our chat room
	loggedUsers := tview.NewTextView()
	loggedUsers.SetBorder(true).SetTitle("Logged Users")

	// window showing all relevant text written in chat
	chat := tview.NewTextView()

	// window showing all logs
	logs := tview.NewTextView()

	chat.SetTitle("Chat").SetBorder(true)
	logs.SetTitle("Logs").SetBorder(true)

	// input line for all user text
	input := tview.NewInputField().
		SetLabel("Me >> ").
		SetFieldWidth(0).
		SetPlaceholder("\\help (ctrl+h) for help menu").
		SetFieldBackgroundColor(tcell.ColorWhite).
		SetFieldTextColor(tcell.ColorBlack)

	chatPage := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(chat, 0, 16, false).
		AddItem(nil, 0, 1, false).
		AddItem(input, 0, 2, true).
		AddItem(logs, 0, 4, false)

	// popup window containing all possible user commands
	helpPage := tview.NewModal().
		SetText(`
\help (ctrl+h) - Show this help
\pm <user> - Private message to <user>
\connect <ip>:<port> - Connect to client <ip>:<port>
\ac - Show all active connections
\dc <user> - Disconnect <user>
\quit - Terminate chat
`).
		AddButtons([]string{"Quit"})

	// name popup to enter user info on startup
	namePage := tview.NewForm().
		AddInputField("Name", "Me", 20, nil, nil).
		AddInputField("Port", "8080", 5, nil, nil).
		AddCheckbox("Age 18+", false, nil)
	namePage.SetBorder(true).SetTitle("Welcome - Enter your name").SetTitleAlign(tview.AlignLeft)

	// name popup to enter user info on startup
	connectPage := tview.NewForm().
		AddInputField("IP", "127.0.0.1", 25, nil, nil).
		AddInputField("Port", "8080", 5, nil, nil)
	connectPage.SetBorder(true).SetTitle("Connect to a new client").SetTitleAlign(tview.AlignLeft)

	// pages object for switching relevant windows as needed ( chat page / help page )
	pages := tview.NewPages().
		AddPage("Chat", chatPage, true, true).
		AddPage("Help", popup(helpPage, 60, 10), true, false).
		AddPage("Name", popup(namePage, 40, 12), true, true).
		AddPage("Connect", popup(connectPage, 40, 10), true, false)

	helpPage.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		pages.HidePage("Help")
		app.SetFocus(input)
	})

	namePage.AddButton("Save", func() {
		name := namePage.GetFormItemByLabel("Name").(*tview.InputField).GetText()
		port := namePage.GetFormItemByLabel("Port").(*tview.InputField).GetText()
		above18 := namePage.GetFormItemByLabel("Age 18+").(*tview.Checkbox).IsChecked()
		localUser.name = name
		localUser.port = port
		localUser.above18 = above18
		pages.HidePage("Name")
		app.SetFocus(input)

		// init all goroutines after welcome screen passed. networking section init
		go serverManager(conns, chat, logs, loggedUsersUpdate)
		go updateLoggedUsers(conns, loggedUsers, logs, loggedUsersUpdate)
	})

	connectPage.AddButton("Connect", func() {
		ip := connectPage.GetFormItemByLabel("IP").(*tview.InputField).GetText()
		port := connectPage.GetFormItemByLabel("Port").(*tview.InputField).GetText()
		if !connectClient(conns, ip+":"+port, chat, logs, loggedUsersUpdate, true) {
			fmt.Fprintf(logs, "Failed connecting to %s:%s\n", ip, port)
		}
		pages.HidePage("Connect")
		app.SetFocus(input)
	})

	// general window for our app
	screen := tview.NewFlex().
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(loggedUsers, 15, 1, false).
			AddItem(tview.NewBox().SetBorder(false).SetTitle("Blank"), 0, 1, false), 15, 1, false).
		AddItem(pages, 0, 3, true)

	input.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			dealWithInput(app, logs, chat, input, pages, helpPage, connectPage, conns)
			input.SetText("")
		}
	})

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Name() {
		// string compare as tcell.KeyCtrlH captured on ctrl+H BUT also on Backspace
		case "Ctrl+Backspace":
			pages.ShowPage("Help")
			app.SetFocus(helpPage)
		}
		return event
	})

	if err := app.SetRoot(screen, true).SetFocus(namePage).Run(); err != nil {
		panic(err)
	}
}

func updateLoggedUsers(conns connections, loggedUsers *tview.TextView, logs *tview.TextView, loggedUsersUpdate chan bool) {
	for {
		select {
		case <-loggedUsersUpdate:
			loggedUsers.Clear()
			for name, _ := range conns {
				if name == "local" {
					continue
				}
				_, err := loggedUsers.Write([]byte(name + "\n"))
				if err != nil {
					fmt.Fprintf(logs, "Failed to show user <%s>: %s", name, err)
				}
			}
		}
	}
}

// function to deal with all user inputs
func dealWithInput(app *tview.Application,
	logs *tview.TextView,
	chat *tview.TextView,
	input *tview.InputField,
	pages *tview.Pages,
	help *tview.Modal,
	connectPage *tview.Form,
	conns connections) {
	text := input.GetText()
	if text == "" {
		return
	}
	switch strings.ToLower(text) {
	case "\\help", "\\h":
		pages.ShowPage("Help")
		app.SetFocus(help)
	case "\\quit", "\\q":
		app.Stop()
	case "\\connect", "\\c":
		pages.ShowPage("Connect")
		app.SetFocus(connectPage)
	case "\\ac":
		fmt.Fprintf(logs, "Active connections:\n--------------\n")
		for _, c := range conns {
			fmt.Fprintf(logs, "%v\n", *c)
		}
		fmt.Fprintf(logs, "--------------\n")
	default:
		splitText := strings.SplitN(text, " ", 3)
		switch strings.ToLower(splitText[0]) {
		case "\\pm":
			if len(splitText) < 3 {
				fmt.Fprintf(logs, "Cant send pm. user / message are missing\n")
				break
			}
			userName := splitText[1]
			text := splitText[2]
			conn := conns[userName]
			if conn == nil {
				fmt.Fprintf(logs, "No such user <%s> to PM\n", userName)
				break
			}
			err := conn.sendMessage("PM@" + text)
			if err != nil {
				fmt.Fprintf(logs, "Failed to send message to %s: %s\n", userName, err)
			} else {
				fmt.Fprintf(chat, "[PM->%s]%s >> %s\n", userName, conns["local"].name, text)
				fmt.Fprintf(logs, "Sending pm to %s\n", userName)

			}
		case "\\dc":
			if len(splitText) != 2 {
				fmt.Fprintf(logs, "Cant disconnect user. \n")
				break
			}
			r := conns[splitText[1]]
			if r == nil {
				fmt.Fprintf(logs, "No such user %s", splitText[1])
				return
			}
			disconnectClient(conns, r, logs)
		default:
			fmt.Fprintf(logs, "Sending broadcast\n")
			// send text to all connected users
			for name, conn := range conns {
				if name == "local" {
					continue
				}
				fmt.Fprintf(logs, "Sending message to %s- %s\n", name, conn.getAddress())
				err := conn.sendMessage(text)
				if err != nil {
					fmt.Fprintf(logs, "Failed to send message to %s: %s\n", name, err)
					continue
				}
			}
			fmt.Fprintf(chat, "%s >> %s\n", conns["local"].name, text)
		}
	}
}
