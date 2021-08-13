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

func (l logger) print(m string) {
	log.Print(l.color + m + "\033[0m")
}

func main() {
	go serverListener()
	time.Sleep(1 * time.Second)
	go openClient()

	time.Sleep(100 * time.Second)

}

func serverListener() {
	fmt.Println("Starting " + connType + " server on " + connHost + ":" + connPort)
	l, err := net.Listen(connType, connHost+":"+connPort)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println("Error connecting:", err.Error())
			return
		}
		fmt.Println("Client " + c.RemoteAddr().String() + " connected.")
		go handleConnection(c)
	}
}

func handleConnection(c net.Conn) {
	l := logger{color: "\033[31m"}
	c_addr := c.RemoteAddr().String()
	for {
		l.print("Server - Waiting for message....")
		b, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Println("Server - Client " + c_addr + " Left.")
			c.Close()
			return
		}
		l.print("Server - Received message from client " + c_addr + " >>> " + b)
		c.Write([]byte("ACK"))
		l.print("Server - Sent ACK to client " + c_addr)
	}
}

func openClient() {
	l := logger{color: "\033[32m"}
	c, err := net.Dial(connType, connHost+":"+connPort)
	if err != nil {
		fmt.Println("Error connecting:", err.Error())
		return
	}
	reader := bufio.NewReader(os.Stdin)

	for {
		l.print("Client - Waiting for text to send: ")
		input, _ := reader.ReadBytes('\n')

		c.Write(input)

		reply := make([]byte, 1024)
		_, err := c.Read(reply)
		if err != nil {
			l.print("Client - Error reading message from server")
		}
		l.print("Client - Server replied >>> " + string(reply))
	}
}
