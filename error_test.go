package fnlang_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xiam/fnlang"
	"github.com/xiam/sexpr/parser"
)

func TestSyntaxError(t *testing.T) {
	testCases := []struct {
		In string
	}{
		{
			In: `(1`,
		},
	}

	for i := range testCases {
		root, err := parser.Parse([]byte(testCases[i].In))
		assert.NoError(t, err)
		assert.NotNil(t, root)
	}
}

func TestUserError(t *testing.T) {
	testCases := []struct {
		In string
	}{
		{
			In: `
							[
								[9 2 7 (:error "failed 1") 5 6 7]
								[8 7 3 4 5]
								[1 (:error "failed 2") 5]
								663
								757
							]
							[8 5 4 2 4]
							5 6
						`,
		},
		{
			In: `
						(defn foo [] [
							(echo "foo")
							1
							[
								9 99
								(:error "stopped")
								999
								[
									3 33 333
								]
							]
							2
							[5 6]
						])
						(foo)
					`,
		},
		{
			In: `
				[
					3
					(defn foo)
					(foo)
				]
				(echo "waka waka")
				(+ 2 5 6)
			`,
		},
	}

	for i := range testCases {
		root, err := parser.Parse([]byte(testCases[i].In))
		assert.NoError(t, err)

		_, result, err := fnlang.Eval(root)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	}
}
