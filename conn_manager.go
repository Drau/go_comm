package main

import (
	"bufio"
	"fmt"
	"github.com/rivo/tview"
	"log"
	"math/rand"
	"net"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	connHost = "127.0.0.1"
	connType = "tcp"
)

type user struct {
	conn    net.Conn
	name    string
	port    string
	above18 bool
	active  bool
}

func (u *user) sendMessage(msg string) error {
	if u.conn != nil {
		_, err := u.conn.Write([]byte(msg + "\n"))
		if err != nil {
			return err
		}
	}
	return nil
}

func (u *user) receiveMessage() (string, error) {
	if u.conn != nil {
		return bufio.NewReader(u.conn).ReadString('\n')
	}
	return "", fmt.Errorf("no open connection")
}

func (u *user) getAddress() string {
	if u.conn != nil {
		if u.port == "" {
			return u.conn.RemoteAddr().String()
		}
		addr := u.conn.RemoteAddr().String()
		splitAddr := strings.Split(addr, ":")
		return splitAddr[0] + ":" + u.port

	}
	return ""

}

func RandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

func removeSpecialChars(s string) string {
	re, err := regexp.Compile("[^a-zA-Z0-9:.]+")
	if err != nil {
		log.Fatal(err)
	}
	return re.ReplaceAllString(s, "")
}

type connections map[string]*user

func serverManager(conns connections, chat *tview.TextView, logs *tview.TextView, loggedUsersUpdate chan bool) {

	connPort := conns["local"].port
	fmt.Fprintf(logs, "Listening for new connections on %s:%s...\n", connHost, connPort)

	l, err := net.Listen(connType, connHost+":"+connPort)
	if err != nil {
		fmt.Fprintf(logs, "Error listening: %s\n", err.Error())
		os.Exit(1)
	}
	defer l.Close()

	for {
		s, err := l.Accept()
		if err != nil {
			fmt.Fprintf(logs, "Error connecting to %s:%s - %s\n", connHost, connPort, err.Error())
			return
		}
		remoteUser := user{name: RandomString(5), above18: true, conn: s, active: true}
		fmt.Fprintf(logs, "Client "+s.RemoteAddr().String()+" connected.\n")
		conns[remoteUser.name] = &remoteUser
		loggedUsersUpdate <- true
		sendDataToNewClient(conns["local"], &remoteUser, logs)
		go handleConnection(conns, &remoteUser, chat, logs, loggedUsersUpdate)
	}
}

func connectClient(conns connections, host string, chat *tview.TextView, logs *tview.TextView, loggedUsersUpdate chan bool, propogateConnect bool) bool {
	if host == "" {
		return false
	}
	// check if connection to client already exists
	for _, c := range conns {
		if c.getAddress() == host {
			fmt.Fprintf(logs, "Already connected to %s", host)
			return true
		}
	}
	s, err := net.Dial(connType, host)
	if err != nil {
		fmt.Fprintf(logs, "Error connecting to <%s>\n%s\n", host, err.Error())
		return false
	}
	remoteUser := user{name: RandomString(5), above18: true, conn: s, active: true}
	_, _ = fmt.Fprintf(logs, "Client (%s) %s connected.\n", remoteUser.name, remoteUser.getAddress())

	conns[remoteUser.name] = &remoteUser
	//loggedUsersUpdate <- true
	sendDataToNewClient(conns["local"], &remoteUser, logs)
	go handleConnection(conns, &remoteUser, chat, logs, loggedUsersUpdate)
	if propogateConnect {time.Sleep(2); retrieveConnList(&remoteUser, logs)}
	return true
}

func handleConnection(conns connections, remoteUser *user, chat *tview.TextView, logs *tview.TextView, loggedUsersUpdate chan bool) {
	fmt.Fprintf(logs, "Handling connection: %s\n", remoteUser.name)
	defer func() {
		delete(conns, remoteUser.name)
		loggedUsersUpdate <- true
	}()
	for {
		message, err := remoteUser.receiveMessage()
		if err != nil {
			fmt.Fprintf(logs, "Connection to %s was closed\n", remoteUser.name)
			return
		}
		fmt.Fprintf(logs, "Received message: %s", message)
		splitMessage := strings.Split(message, "@")
		switch len(splitMessage) {
		// regular message
		case 1:
			fmt.Fprintf(chat, "%s >> %s\n", remoteUser.name, strings.TrimSuffix(message, "\n"))
		// command from client
		case 2:
			switch splitMessage[0] {
			case "CONNECT":
				ips := strings.Split(splitMessage[1], "|")
				for _, ip := range ips {
					if len(ip) == 1 {continue}
					connectClient(conns, removeSpecialChars(ip), chat, logs, loggedUsersUpdate, false)
				}
			case "INFO":
				data := strings.Split(splitMessage[1], "|")
				oldName := remoteUser.name
				remoteUser.name = data[0]
				remoteUser.port = removeSpecialChars(data[1])
				conns[remoteUser.name] = remoteUser
				delete(conns, oldName)
				loggedUsersUpdate <- true
			case "PM":
				fmt.Fprintf(chat, "[PM]%s >> %s\n", remoteUser.name, strings.TrimSuffix(splitMessage[1], "\n"))
			case "GETCONNS":
				sendConnectionList(conns, remoteUser, logs)
			}


		}

	}
}

func sendDataToNewClient(localUser *user, remoteUser *user, logs *tview.TextView) {
	msg := []string{localUser.name, localUser.port}
	final := "INFO@" + strings.Join(msg, "|")
	err := remoteUser.sendMessage(final)
	if err != nil {
		fmt.Fprintf(logs, "Failed to send data to new clien: %s", err)
		return
	}
	fmt.Fprintf(logs, "Sent INFO message to <%s>\n", remoteUser.name)
}

func disconnectClient(conns connections, remoteUser *user, logs *tview.TextView) {
	remoteUser.conn.Close()
	final := "DISCONNECT@" + remoteUser.name
	for name, conn := range conns {
		if name != "local" && name != remoteUser.name {
			err := conn.sendMessage(final)
			if err != nil {
				fmt.Fprintf(logs, "Failed to send DISCONNECT message: %s", err)
				return
			}
			fmt.Fprintf(logs, "Sent DISCONNECT message to <%s>\n", conn.name)
		}
	}
}

func retrieveConnList(remoteUser *user, logs *tview.TextView) {
	final := "GETCONNS@"
	err := remoteUser.sendMessage(final)
	if err != nil {
		fmt.Fprintf(logs, "Failed to request connection list from new client: %s", err)
		return
	}
	fmt.Fprintf(logs, "Sent GETCONNS message to <%s>\n", remoteUser.name)
}

func sendConnectionList(conns connections, remoteUser *user, logs *tview.TextView) {
	msg := make([]string, len(conns) -2)
	for name, c := range conns {
		if name != remoteUser.name && name != "local" {msg = append(msg, c.getAddress())}
	}
	//msg = append("" + connPort)
	final := "CONNECT@" + strings.Join(msg, "|")
	err := remoteUser.sendMessage(final)
	if err != nil {
		fmt.Fprintf(logs, "Failed to send connection list to new client: %s", err)
		return
	}
	fmt.Fprintf(logs, "Sent CONNECT message to <%s>\n", remoteUser.name)
}

