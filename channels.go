package main

import (
	"sync"
)

type Files struct {
	Chan  chan File
	Close sync.Once
}

func (c *Files) close() {
	c.Close.Do(func() {
		close(c.Chan)
	})
}

func (c *Files) empty() {
	for {
		select {
		case _, ok := <-c.Chan:
			if !ok {
				return
			}
		default:
			return
		}
	}
}

type Chunks struct {
	Chan  chan Chunk
	Close sync.Once
}

func (c *Chunks) close() {
	c.Close.Do(func() {
		close(c.Chan)
	})
}

func (c *Chunks) empty() {
	for {
		select {
		case _, ok := <-c.Chan:
			if !ok {
				return
			}
		default:
			return
		}
	}
}

type Articles struct {
	Chan  chan Article
	Close sync.Once
}

func (c *Articles) close() {
	c.Close.Do(func() {
		close(c.Chan)
	})
}

func (c *Articles) empty() {
	for {
		select {
		case _, ok := <-c.Chan:
			if !ok {
				return
			}
		default:
			return
		}
	}
}

var (
	files        Files
	inputChunks  Chunks
	outputChunks Chunks
	articles     Articles
)

func init() {
	// make channels
	files.Chan = make(chan File, 1)
	inputChunks.Chan = make(chan Chunk, conf.Connections*2)
	outputChunks.Chan = make(chan Chunk, conf.Connections*2)
	articles.Chan = make(chan Article, conf.Connections*2)
}
