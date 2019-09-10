package signal

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrapErrorWithCodeReturnsCode(t *testing.T) {
	e := WrapErrorWithCode(io.EOF, 42)
	ee, ok := e.(errorWithExitCode)
	assert.True(t, ok)
	assert.Equal(t, ee.exitCode, 42)
}

func TestWrapErrorWithCodeReturnsNilWhenNilGiven(t *testing.T) {
	assert.Nil(t, WrapErrorWithCode(nil, 42))
}

func TestGetCodeFromErrorReturnsExitCode(t *testing.T) {
	e := WrapErrorWithCode(io.EOF, 42)
	assert.Equal(t, 42, getCodeFromError(e, -1))
}

func TestGetCodeFromErrorReturnsDefaultValue(t *testing.T) {
	assert.Equal(t, 42, getCodeFromError(io.EOF, 42))
}

func TestNewTerminationFuncRunsThenReturnsNil(t *testing.T) {
	called := 0
	fn := NewTerminationFunc(func() { called++ })
	assert.Nil(t, fn(nil))
	assert.Equal(t, 1, called)
}
