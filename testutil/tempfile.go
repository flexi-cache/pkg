package testutil

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// WithTempFile creates a tempfile with given content then calls f()
//
// The file will be deleted after calling f()
func WithTempFile(t *testing.T, content string, f func(filename string)) {
	tmpfile, err := ioutil.TempFile("", t.Name())
	assert.NoError(t, err)

	defer func() {
		if tmpfile != nil {
			tmpfile.Close()
		}
		os.Remove(tmpfile.Name())
	}()

	_, err = tmpfile.Write([]byte(content))
	assert.NoError(t, err)

	f(tmpfile.Name())
}
