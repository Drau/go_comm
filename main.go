package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)

const (
	connHost = "127.0.0.1"
	connPort = "8080"
	connType = "tcp"
)

func main() {
	ipFlag := flag.String("i", "", "help message for flag n")
	portFlag := flag.String("p", "", "help message for flag n")
	flag.Parse()
	conns := make(connections, 0, 100)
	if !openClient(&conns, *ipFlag, *portFlag) {
		go serverManager(&conns)
	}

	go consoleListener(&conns)

	time.Sleep(100 * time.Second)

}

func openClient(conns *connections, ip string, port string) bool {
	if ip == "" || port == "" {
		return false
	}
	s, err := net.Dial(connType, connHost+":"+connPort)
	if err != nil {
		fmt.Println("Error connecting:", err.Error())
		return false
	}
	session_P := newSessionConn(s)
	fmt.Println("Client " + session_P.getAddress() + " connected.")

	(*conns) = append((*conns), session_P)

	go handleConnection(conns, session_P)
	return true
}

func serverManager(conns *connections) {
	fmt.Println("Starting " + connType + " server on " + connHost + ":" + connPort)

	l, err := net.Listen(connType, connHost+":"+connPort)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer l.Close()

	for {
		s, err := l.Accept()
		if err != nil {
			fmt.Println("Error connecting:", err.Error())
			return
		}
		session_P := newSessionConn(s)
		fmt.Println("Client " + session_P.getAddress() + " connected.")

		(*conns) = append((*conns), session_P)

		go handleConnection(conns, session_P)
	}
}

func handleConnection(conns *connections, s *session_conn) {
	l := logger{color: "\033[31m"}
	ch := make(chan string)

	//goroutine to wait(block) for input on connection, then pass it back to the conenction handler
	go func(s *session_conn, ch chan string) {
		for {
			b, err := bufio.NewReader((*s).conn).ReadString('\n')
			if err != nil {
				b = "/quit"
				ch <- b
				return
			}
			ch <- b

		}
	}(s, ch)

	for {
		select {
		case b := <-ch:
			if b == "/quit" {
				fmt.Println("Client " + (*s).getAddress() + " Left.")
				(*s).close()
				return
			}
			l.print("Message recieved from " + (*s).getAddress() + " >>> " + b)
		case cmd := <-(*s).cmd_chan:
			(*s).conn.Write([]byte(cmd))
		}
	}
}

func consoleListener(conns *connections) {
	reader := bufio.NewReader(os.Stdin)
	for {
		input, _ := reader.ReadString('\n')
		found := false
		for _, s := range *conns {
			if s.active {
				s.cmd_chan <- input
				found = true
			}
		}
		if !found {
			fmt.Println("No available clients to recieve the message")
		}
	}
}

// type i_message struct {
// 	message string
// 	cmd     string
// 	to      *[]session_conn
// }

// func parseRawInput(conns *connections, i string) {

// }
