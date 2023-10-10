package csvstutter

import (
	"bytes"
	"encoding/csv"
	"io"
	"strings"
	"sync"
)

type Reader struct {
	err      error
	reader   *csv.Reader
	done     chan struct{}
	lock     sync.Mutex
	toBeRead chan *bytes.Buffer
}

func NewReader(rdr io.Reader) *Reader {
	reader := csv.NewReader(rdr)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1 // disabling this check
	toUnstutter := make(chan []string, 100)
	toWrite := make(chan []string, 100)

	r := &Reader{
		reader:   reader,
		done:     make(chan struct{}),
		toBeRead: make(chan *bytes.Buffer),
	}
	toWriteCopy := toWrite // create local variable copy for 3rd goroutine later
	go func() {
		// read lines from rdr
		defer func() {
			close(toUnstutter)
		}()
		for {
			line, err := reader.Read()
			if err == io.EOF {
				return
			}
			if err != nil {
				r.lock.Lock()
				r.err = err
				r.lock.Unlock()
				return
			}
			select {
			case <-r.done:
				return
			case toUnstutter <- line:
			}
		}
	}()
	go func() {
		// unstutter them!
		defer func() {
			close(toWrite)
		}()
		for line := range toUnstutter {
			for x := range line {
				idx := strings.Index(line[x], "\n")
				if idx == -1 {
					continue // no mulitple lines here
				}
				leftIdx := idx
				if idx > 0 && line[x][idx-1] == '}' {
					leftIdx--
				}
				if line[x][:leftIdx] != line[x][idx+1:] {
					continue // not a stutter, not the same
				}
				// yes, we're a stutter value! remove it
				line[x] = line[x][:leftIdx]
			}
			select {
			case <-r.done:
				return
			case toWrite <- line:
				// log.Printf("sent line toWrite")
			}
		}
	}()
	go func() {
		// write it!
		buf := &bytes.Buffer{}
		csvwriter := csv.NewWriter(buf)
		doneReading := false
		allowReading := (chan *bytes.Buffer)(nil)
		for {
			select {
			case line, ok := <-toWriteCopy:
				if !ok { // toWrite is closed
					doneReading = true
					toWriteCopy = nil
					// log.Printf("toWrite is closed")
					continue
				}
				csvwriter.Write(line)
				csvwriter.Flush()
				allowReading = r.toBeRead // only now should buf be sent to read()
				// log.Printf("allowReading set to r.toBeRead")
			case allowReading <- buf:
				// log.Printf("sent buf to read()")
				<-allowReading // get it back from the read function
				// log.Printf("received buf from read()")
			case <-r.done:
			}
			if doneReading && buf.Len() == 0 {
				close(allowReading)
				return
			}
		}
	}()
	return r
}

// Close cleans up the resources created to read the file multicore style
func (reader *Reader) Close() {
	close(reader.done)
}

// Read returns only valid CSV data as read from the source and removes multilines and stutter
// if there's an error all subsequent calls to Read will fail with the same error
func (reader *Reader) Read(lineIn []byte) (int, error) {
	reader.lock.Lock()
	err := reader.err
	reader.lock.Unlock()
	if err != nil {
		return 0, err
	}
	// log.Printf("read() called")
	buf, ok := <-reader.toBeRead
	// log.Printf("read got the buffer")
	if !ok {
		// log.Printf("EOF in Read() call")
		return 0, io.EOF
	}
	num, err := buf.Read(lineIn)
	reader.toBeRead <- nil
	if num == 0 && err == io.EOF {
		return reader.Read(lineIn)
	}
	// log.Printf("read() returned %d for %v", num, err)
	return num, err
}
