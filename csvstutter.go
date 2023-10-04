package csvstutter

import (
	"bytes"
	"encoding/csv"
	"io"
	"strings"
)

type Reader struct {
	err      error
	reader   *csv.Reader
	toReturn *bytes.Buffer
	writer   *csv.Writer
}

func NewReader(rdr io.Reader) *Reader {
	buffer := &bytes.Buffer{}
	reader := csv.NewReader(rdr)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1 // disabling this check
	return &Reader{
		reader:   reader,
		writer:   csv.NewWriter(buffer),
		toReturn: buffer,
	}
}

// Read returns only valid CSV data as read from the source and removes multilines and stutter
// if there's an error all subsequent calls to Read will fail with the same error
func (reader Reader) Read(lineIn []byte) (int, error) {
	if reader.err != nil {
		return 0, reader.err
	}
	if reader.toReturn.Len() == 0 { // nothing to copy, need another line loaded
		line, err := reader.reader.Read()
		if err != nil {
			reader.err = err
			return 0, err
		}
		for x := range line {
			idx := strings.Index(line[x], "\n")
			if idx == -1 {
				continue // no mulitple lines here
			}
			if line[x][:idx] != line[x][idx+1:] {
				continue // not a stutter, not the same
			}
			// yes, we're a stutter value! remove it
			line[x] = line[x][:idx]
		}
		err = reader.writer.Write(line)
		if err != nil {
			reader.err = err
			return 0, err
		}
		reader.writer.Flush()
	}
	return reader.toReturn.Read(lineIn)
}
