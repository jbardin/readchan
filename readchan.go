package readchan

import (
	"bufio"
	"io"
	"log"
	"sync"
)

type Chunk struct {
	Data []byte
	Err  error

	pool *sync.Pool
}

func (c *Chunk) Done() {
	c.pool.Put(c)
}

func Chunks(r io.Reader, chunkSize int, cancel chan bool) <-chan *Chunk {
	if chunkSize <= 0 {
		panic("invalid chunk size")
	}

	pool := sync.Pool{}
	pool.New = func() interface{} {
		return &Chunk{
			Data: make([]byte, chunkSize),
			pool: &pool,
		}
	}

	readChan := make(chan *Chunk)

	go func() {
		var (
			n   int
			err error
		)

		defer close(readChan)

		for {
			select {
			case <-cancel:
				log.Println("ReadChan canceled")
				return
			default:
			}

			chunk := pool.Get().(*Chunk)

			n, err = r.Read(chunk.Data)
			chunk.Data = chunk.Data[:n]
			chunk.Err = err
			readChan <- chunk

			if err != nil {
				return
			}
		}
	}()

	return readChan
}

func Lines(r io.Reader, chunkSize int, cancel chan bool) <-chan *Chunk {
	if chunkSize <= 0 {
		panic("invalid chunk size")
	}

	pool := sync.Pool{}
	pool.New = func() interface{} {
		return &Chunk{
			Data: make([]byte, 0, chunkSize),
			pool: &pool,
		}
	}

	readChan := make(chan *Chunk)

	scanner := bufio.NewScanner(r)

	go func() {
		defer close(readChan)

		for scanner.Scan() {
			select {
			case <-cancel:
				break
			default:
			}

			chunk := (pool.Get().(*Chunk))
			chunk.Data = chunk.Data[:0]
			chunk.Data = append(chunk.Data, scanner.Bytes()...)
			chunk.Err = nil

			readChan <- chunk
		}

		if err := scanner.Err(); err != nil {
			chunk := (pool.Get().(*Chunk))
			chunk.Data = nil
			chunk.Err = err

			readChan <- chunk
		}

	}()

	return readChan

}
