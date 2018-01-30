// Package readchan provides methods for interating over an io.Reader by block
// or line, and reading the results via a channel.
package readchan

import (
	"bufio"
	"context"
	"io"
	"sync"
)

var (
	MaxScanTokenSize = bufio.MaxScanTokenSize
)

// A Chunk contains the []byte Data and error from the previous Read operation.
// The Data []byte slice is safe for local use until Done is called returning
// the Chunk to the pool.
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
// The maxSize argument sets the allocated capacity of each []byte. Reads will
// buffer readAhead number of Chunks in the channel as soon as they are
// available.  Canceling the context will cause the Reads loop to return, but
// it cannot interrupt pending Read calls on r.
func Reads(ctx context.Context, r io.Reader, maxSize, readAhead int) <-chan *Chunk {
	if maxSize <= 0 {
		panic("invalid max buffer size")
	}

	if readAhead < 0 {
		readAhead = 1
	}

	pool := sync.Pool{}
	pool.New = func() interface{} {
		return &Chunk{
			Data: make([]byte, maxSize),
			pool: &pool,
		}
	}

	readChan := make(chan *Chunk, readAhead)

	go func() {
		var n int
		var err error

		defer close(readChan)

		for {
			chunk := pool.Get().(*Chunk)

			n, err = r.Read(chunk.Data)
			chunk.Data = chunk.Data[:n]
			chunk.Err = err

			select {
			case readChan <- chunk:
			case <-ctx.Done():
				return
			}

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
// The readAhead argument determines the buffer size for the channel, which
// will be filled as soon as data available. Canceling the context will
// cause the Lines scanner loop to return, but it cannot interrupt pending Read
// calls on r.
func Lines(ctx context.Context, r io.Reader, readAhead int) <-chan *Chunk {
	if readAhead < 0 {
		readAhead = 1
	}

	pool := sync.Pool{}
	pool.New = func() interface{} {
		return &Chunk{
			pool: &pool,
		}
	}

	readChan := make(chan *Chunk, readAhead)

	scanner := bufio.NewScanner(r)

	go func() {
		defer close(readChan)

		for scanner.Scan() {
			chunk := (pool.Get().(*Chunk))
			if chunk.Data != nil {
				chunk.Data = chunk.Data[:0]
			}
			chunk.Err = nil

			chunk.Data = append(chunk.Data, scanner.Bytes()...)

			select {
			case readChan <- chunk:
			case <-ctx.Done():
				return
			}
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
