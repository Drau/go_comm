package main

import "net"

type session_conn struct {
	conn     net.Conn
	cmd_chan chan string
	name     string
	active   bool
}

func newSessionConn(c net.Conn) session_conn {
	return session_conn{c, make(chan string), "", true}
}

func (s *session_conn) getAddress() string {
	return s.conn.RemoteAddr().String()
}

func (s *session_conn) close() {
	s.conn.Close()
	s.active = false
}

type connections []session_conn
