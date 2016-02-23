package readchan

import (
	"bufio"
	"io"
	"sync"
)

var (
	MaxScanTokenSize = bufio.MaxScanTokenSize
)

// A Chunk contains the []byte Data and error from the previous Read operation.
type Chunk struct {
	Data []byte
	Err  error

	pool *sync.Pool
}

// Done returns the Chunk to a pool to be reused for subsequent Reads.
func (c *Chunk) Done() {
	c.pool.Put(c)
}

// Reads returns a channel that will send a Chunk for every Read on r.
//
// The maxSize argument sets the allocated capacity of each []byte.  Closing
// the cancel channel will cause Reads to return after the next Read operation.
func Reads(r io.Reader, maxSize int, cancel chan bool) <-chan *Chunk {
	if maxSize <= 0 {
		panic("invalid max buffer size")
	}

	pool := sync.Pool{}
	pool.New = func() interface{} {
		return &Chunk{
			Data: make([]byte, maxSize),
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

// Lines returns a channel that will send a Chunk for every line read from r.
// Lines are read via a bufio.Scanner, and do not include the newline
// characters.
//
// The minSize argument sets the minimum allocated size of each []byte, which
// may be extended to accommodate longer lines. Closing the cancel channel
// will cause the Lines reader to return after the next Read operation.
func Lines(r io.Reader, minSize int, cancel chan bool) <-chan *Chunk {
	if minSize < 0 {
		panic("invalid min buffer size")
	}

	pool := sync.Pool{}
	pool.New = func() interface{} {
		return &Chunk{
			Data: make([]byte, 0, minSize),
			pool: &pool,
		}
	}

	readChan := make(chan *Chunk)

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, minSize), MaxScanTokenSize)

	go func() {
		defer close(readChan)

		for scanner.Scan() {
			select {
			case <-cancel:
				return
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
