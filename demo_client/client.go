package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

const (
	connHost = "127.0.0.1"
	connPort = "8080"
	connType = "tcp"
)

type logger struct {
	color string
}

func (l logger) print(m ...string) {
	message := ""
	for _, v := range m {
		message += v
	}
	log.Print(l.color + message + "\033[0m")
}

func main() {

	time.Sleep(1 * time.Second)
	go openClient()

	time.Sleep(100 * time.Second)

}

func openClient() {
	l := logger{color: "\033[32m"}
	c, err := net.Dial(connType, connHost+":"+connPort)
	if err != nil {
		fmt.Println("Error connecting:", err.Error())
		return
	}
	reader := bufio.NewReader(os.Stdin)

	go func(c *net.Conn) {
		for {
			reply := make([]byte, 1024)
			_, err := (*c).Read(reply)
			if err != nil {
				l.print("Error reading message from server")
				os.Exit(1)
			}
			l.print("Message recieved >>> " + string(reply))
		}
	}(&c)

	for {
		input, _ := reader.ReadBytes('\n')

		c.Write(input)
	}
}
