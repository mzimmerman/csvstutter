package csvstutter_test

import (
	"bytes"
	"encoding/csv"
	"io/ioutil"
	"log"
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

	got, err := ioutil.ReadAll(csvstutter.NewReader(strings.NewReader(data)))
	if err != nil {
		t.Errorf("got unexpected error reading - %v", err)
	}
	if expected != string(got) {
		t.Errorf("\nexpected: %q\ngot     : %q", expected, got)
	}
}

var bigResults = [][]string(nil)

func BenchmarkStutter(b *testing.B) {
	dataIn, err := os.ReadFile("data.out")
	if err != nil {
		log.Fatalf("error reading in file for testing")
	}
	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		stutter := csvstutter.NewReader(bytes.NewReader(dataIn))
		results, err := csv.NewReader(stutter).ReadAll()
		if err != nil {
			b.Errorf("error reading csv - %v", err)
		}
		bigResults = results
	}
}
