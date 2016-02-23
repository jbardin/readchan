package readchan

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

func TestChunkReader(t *testing.T) {
	orig, err := ioutil.ReadFile("./testdata/test.data")
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.Open("./testdata/test.data")
	if err != nil {
		t.Fatal(err)
	}

	var newData []byte

	readChan := Reads(f, 1024, nil)
	for chunk := range readChan {
		newData = append(newData, chunk.Data...)
		chunk.Done()
		if chunk.Err != nil && chunk.Err != io.EOF {
			t.Fatal(chunk.Err)
		}
	}

	if !bytes.Equal(orig, newData) {
		t.Fatal("mismatched data")
	}

}

func TestLineReader(t *testing.T) {
	orig, err := ioutil.ReadFile("./testdata/test.data")
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.Open("./testdata/test.data")
	if err != nil {
		t.Fatal(err)
	}

	var newData []byte

	lines := 0
	readChan := Lines(f, 1024, nil)
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

	if !bytes.Equal(orig, newData) {
		t.Fatal("mismatched data")
	}
}

func BenchmarkChunkReader(b *testing.B) {
	f, err := os.Open("./testdata/test.data")
	if err != nil {
		b.Fatal(err)
	}

	newData := make([]byte, 0)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		newData = newData[:0]
		f.Seek(0, 0)
		readChan := Reads(f, 1024, nil)

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
	f, err := os.Open("./testdata/test.data")
	if err != nil {
		b.Fatal(err)
	}

	newData := make([]byte, 0)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		newData = newData[:0]
		f.Seek(0, 0)
		readChan := Lines(f, 512, nil)
		for chunk := range readChan {
			newData = append(newData, chunk.Data...)
			chunk.Done()
			if chunk.Err != nil && chunk.Err != io.EOF {
				b.Fatal(chunk.Err)
			}
		}
	}
}
