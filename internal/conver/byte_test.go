package conver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBytesToString(t *testing.T) {
	b := []byte("hello")
	s := BytesToString(b)
	assert.Equal(t, "hello", s)
}
