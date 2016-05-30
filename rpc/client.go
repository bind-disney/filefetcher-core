package rpc

import "net"

type client struct {
	directory  string
	connection *net.Conn
}

func (c *client) Address() string {
	clientConnection := *c.connection

	return clientConnection.RemoteAddr().String()
}
