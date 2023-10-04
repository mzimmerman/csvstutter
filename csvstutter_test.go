package csvstutter_test

import (
	"io/ioutil"
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
