package conver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringToBytes(t *testing.T) {
	s := "hello"
	b := StringToBytes(s)
	assert.Equal(t, []byte("hello"), b)
}
