package readchan

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

var (
	testFile = "./testdata/test.data.gz"

	testData   []byte
	testReader io.ReadSeeker
)

func init() {
	var err error
	f, err := os.Open(testFile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	r, err := gzip.NewReader(f)
	if err != nil {
		log.Fatal(err)
	}

	testData, err = ioutil.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}

	testReader = bytes.NewReader(testData)
}

func TestChunkReader(t *testing.T) {
	defer testReader.Seek(0, 0)

	var newData []byte

	readChan := Reads(testReader, 1024, 1, nil)
	for chunk := range readChan {
		newData = append(newData, chunk.Data...)
		chunk.Done()
		if chunk.Err != nil && chunk.Err != io.EOF {
			t.Fatal(chunk.Err)
		}
	}

	if !bytes.Equal(testData, newData) {
		t.Fatal("mismatched data")
	}

}

func TestLineReader(t *testing.T) {
	defer testReader.Seek(0, 0)

	var newData []byte

	lines := 0
	readChan := Lines(testReader, 1024, 1, nil)
	for chunk := range readChan {
		newData = append(newData, chunk.Data...)
		newData = append(newData, '\n')
		lines++

		chunk.Done()
		if chunk.Err != nil && chunk.Err != io.EOF {
			t.Fatal(chunk.Err)
		}
	}

	if lines != 1140 {
		t.Fatalf("incorrect line count. Counted %d, expected %d", lines, 1140)
	}

	if !bytes.Equal(testData, newData) {
		t.Fatal("mismatched data")
	}
}

func BenchmarkChunkReader(b *testing.B) {
	defer testReader.Seek(0, 0)

	newData := make([]byte, 0)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		newData = newData[:0]
		testReader.Seek(0, 0)
		readChan := Reads(testReader, 1024, 1, nil)

		for chunk := range readChan {
			newData = append(newData, chunk.Data...)
			chunk.Done()
			if chunk.Err != nil && chunk.Err != io.EOF {
				b.Fatal(chunk.Err)
			}
		}

	}
}

func BenchmarkLineReader(b *testing.B) {
	defer testReader.Seek(0, 0)

	newData := make([]byte, 0)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		newData = newData[:0]
		testReader.Seek(0, 0)
		readChan := Lines(testReader, 512, 1, nil)
		for chunk := range readChan {
			newData = append(newData, chunk.Data...)
			chunk.Done()
			if chunk.Err != nil && chunk.Err != io.EOF {
				b.Fatal(chunk.Err)
			}
		}
	}
}
