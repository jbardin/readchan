# readchan 

[![GoDoc](https://godoc.org/github.com/jbardin/readchan?status.svg)](https://godoc.org/github.com/jbardin/readchan)

Package readchan provides methods for interating over an `io.Reader` by block
or line, and reading the results via a channel.


    f, err := os.Open("file.txt")
    if err != nil {
      log.Fatal(err)
    }

    for line := range readchan.Lines(f, 1, nil) {
      if line.Err != nil {
        log.Fatal(err)
      }
      fmt.Printf("LINE: %q\n", line.Data)
	  line.Done()
    }
