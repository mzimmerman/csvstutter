package csvstutter

import (
	"bytes"
	"encoding/csv"
	"io"
	"strings"
)

type Reader struct {
	reader   *csv.Reader
	done     chan error
	toBeRead *bytes.Buffer
	toWrite  chan []string
	writer   *csv.Writer
}

func NewReader(rdr io.Reader) *Reader {
	reader := csv.NewReader(rdr)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1 // disabling this check
	buf := &bytes.Buffer{}
	r := &Reader{
		reader:   reader,
		done:     make(chan error),
		toBeRead: buf,
		toWrite:  make(chan []string, 100),
		writer:   csv.NewWriter(buf),
	}
	go func() {
		// read and unstutter them!
		defer func() {
			close(r.toWrite)
		}()
		for {
			line, err := reader.Read()
			if err == io.EOF {
				return
			}
			if err != nil {
				select {
				case _, ok := <-r.done:
					_ = ok // we don't use this but I need to receive a value for the compiler
					// done processing
				case r.done <- err:
					// error was sent, stop reading
				}
				return
			}
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
			case _, ok := <-r.done:
				_ = ok // we don't use this but I need to receive a value for the compiler
				// done processing
			case r.toWrite <- line:
				// log.Printf("sent line toWrite")
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
	if reader.toBeRead.Len() == 0 {
		select {
		case line, ok := <-reader.toWrite:
			if !ok {
				return 0, io.EOF
			}
			reader.writer.Write(line)
			reader.writer.Flush()
		case err := <-reader.done:
			return 0, err
		}
	}
	return reader.toBeRead.Read(lineIn)
}
