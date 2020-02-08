package fnlang

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xiam/sexpr/parser"
)

func TestErrorMessages(t *testing.T) {
	testCases := []struct {
		In  string
		Out string
	}{
		{
			In:  `(1`,
			Out: `[1]`,
		},
	}

	for i := range testCases {
		root, err := parser.Parse([]byte(testCases[i].In))
		assert.Error(t, err)
		assert.Nil(t, root)
	}
}
