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
	session := newSessionConn(s)
	fmt.Println("Client " + session.getAddress() + " connected.")

	(*conns) = append((*conns), session)

	go handleConnection(conns, &session)
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
		session := newSessionConn(s)
		fmt.Println("Client " + session.getAddress() + " connected.")

		(*conns) = append((*conns), session)

		go handleConnection(conns, &session)
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
		for _, s := range *conns {
			if s.active {
				fmt.Println(">>> 1: " + input)
				s.cmd_chan <- input
				Im never getting here after client closes remote session !!!!! maybe use chanels to pass info about closed connection to main() where it will save all conns
				fmt.Println(">>>  2")
			}
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
