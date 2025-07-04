package multimatch_test

import (
	"iter"
	"testing"

	"github.com/xoltia/botsu-oshi-stats/multimatch"
)

func TestMatcher(t *testing.T) {
	builder := multimatch.Builder{}
	builder.AddString("meat", 1)
	builder.AddString("meet", 2)
	builder.AddString("eat", 3)
	builder.AddString("eating", 4)
	builder.AddString("tiny", 5)
	builder.AddString("in", 6)
	matcher := builder.Build()

	expected := []int{3, 6, 4, 1, 3}
	output := matcher.SearchString("I am eating meat")
	next, _ := iter.Pull(output)
	for i, e := range expected {
		actual, ok := next()
		if !ok {
			break
		}
		if e != actual {
			t.Errorf("Expected %d got %d at %d", e, actual, i)
		}
	}
}
