package main

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/Tensai75/nntp"
)

type safeConn struct {
	sync.RWMutex
	closed bool
	*nntp.Conn
}

var (
	initConnGuard   sync.Once
	connectionGuard chan struct{}
)

func ConnectNNTP() (*safeConn, error) {
	initConnGuard.Do(func() {
		connectionGuard = make(chan struct{}, conf.Connections)
	})
	connectionGuard <- struct{}{} // will block if guard channel is already filled
	var conn *nntp.Conn
	if conf.SSL {
		conn, err = nntp.DialTLS("tcp", conf.Host+":"+strconv.Itoa(conf.Port), nil)
	} else {
		conn, err = nntp.Dial("tcp", conf.Host+":"+strconv.Itoa(conf.Port))
	}
	if err != nil {
		return nil, fmt.Errorf("Connection to usenet server failed: %v\r\n", err)
	}
	safeConn := safeConn{
		Conn: conn,
	}
	if err = safeConn.Authenticate(conf.NntpUser, conf.NntpPass); err != nil {
		safeConn.Close()
		return nil, fmt.Errorf("Authentication with usenet server failed: %v\r\n", err)
	}
	return &safeConn, nil
}

func (c *safeConn) Close() {
	c.Lock()
	defer c.Unlock()
	if !c.closed {
		if c.Conn != nil {
			c.Quit()
		}
		if len(connectionGuard) > 0 {
			<-connectionGuard
		}
		c.closed = true
	}
}
