package main

var (
	filesChan        chan File
	inputChunksChan  chan Chunk
	outputChunksChan chan Chunk
	articlesChan     chan Article
	headerCheckChan  chan Article
)

func init() {
	// make channels
	filesChan = make(chan File, 1)
	inputChunksChan = make(chan Chunk, conf.Connections*4)
	outputChunksChan = make(chan Chunk, conf.Connections*4)
	articlesChan = make(chan Article, conf.Connections*4)
	headerCheckChan = make(chan Article, conf.Connections*4*2)
}
