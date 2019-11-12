package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAssembleDir(t *testing.T) {
	assert.Equal(t, "/", AssembleDir("/", ""))
	assert.Equal(t, "/abc", AssembleDir("/", "abc"))
	assert.Equal(t, "/asd/asdc/z", AssembleDir("/asd/", "/asdc/z"))
	assert.Equal(t, "/", AssembleDir("", "/"))
}
