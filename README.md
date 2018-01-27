# readchan 

[![GoDoc](https://godoc.org/github.com/jbardin/readchan?status.svg)](https://godoc.org/github.com/jbardin/readchan)

Package readchan provides methods for iterating over an `io.Reader` by block
or line, and reading the results via a channel.


    f, err := os.Open("file.txt")
    if err != nil {
      log.Fatal(err)
    }

    for line := range readchan.Lines(context.TODO(), f, 1) {
      if line.Err != nil {
        log.Fatal(err)
      }
      fmt.Printf("LINE: %q\n", line.Data)
	  line.Done()
    }
