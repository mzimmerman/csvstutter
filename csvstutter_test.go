package csvstutter_test

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"
	"testing"

	"github.com/mzimmerman/csvstutter"
)

func TestStutter(t *testing.T) {
	data := `hi,there,"how
how",are,you,doing?,"i am great
thank you"
another line
csv,is,fun`

	expected := `hi,there,how,are,you,doing?,"i am great
thank you"
another line
csv,is,fun
`

	got, err := io.ReadAll(csvstutter.NewReader(strings.NewReader(data), 10))
	if err != nil {
		t.Errorf("got unexpected error reading - %v", err)
	}
	if expected != string(got) {
		t.Errorf("\nexpected: %q\ngot     : %q", expected, got)
	}
}

var bigResults = [][]string{}

func BenchmarkStutter1(b *testing.B) {
	benchmarkStutter(b, 1)
}

func BenchmarkStutter2(b *testing.B) {
	benchmarkStutter(b, 2)
}

func BenchmarkStutter10(b *testing.B) {
	benchmarkStutter(b, 10)
}

func BenchmarkStutter100(b *testing.B) {
	benchmarkStutter(b, 100)
}

func BenchmarkStutter1000(b *testing.B) {
	benchmarkStutter(b, 1000)
}

func benchmarkStutter(b *testing.B, size int) {
	dataIn, err := os.ReadFile("data.out")
	if err != nil {
		log.Fatalf("error reading in file for testing")
	}
	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		stutter := csvstutter.NewReader(bytes.NewReader(dataIn), 10)
		results, err := csv.NewReader(stutter).ReadAll()
		if err != nil {
			b.Errorf("error reading csv - %v", err)
		}
		stutter.Close()
		bigResults = results
	}
}

// generateData makes a csvstutter file for testing
func generateData() {
	chars := `qwertyuiopasdfghjklzxcvbnm,"'1234567890`
	f, err := os.OpenFile("data.out", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		log.Fatalf("error opening file - %v", err)
	}
	defer f.Close()
	var field bytes.Buffer
	var line []string
	var stutterLine []string
	writer := csv.NewWriter(f)
	defer writer.Flush()
	for x := 0; x < 50000; x++ {
		line = line[:0]
		stutterLine = stutterLine[:0]
		for y := 0; y < 25; y++ {
			length := rand.Intn(200) + 25
			field.Reset()
			for z := 0; z < length; z++ {
				field.WriteByte(chars[rand.Intn(len(chars))])
			}
			line = append(line, field.String())
			stutterLine = append(stutterLine, fmt.Sprintf("%s\n%s", field.Bytes(), field.Bytes()))
		}
		err = writer.Write(line)
		if err != nil {
			log.Fatalf("error writing line - %v", err)
		}
		err = writer.Write(stutterLine)
		if err != nil {
			log.Fatalf("error writing stutterline - %v", err)
		}
	}
}
