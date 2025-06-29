package multimatch_test

import (
	"fmt"
	"testing"

	"github.com/xoltia/botsu-oshi-stats/internal/multimatch"
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

	output := matcher.SearchString("I am eating meat")
	for o := range output {
		fmt.Printf("Output: %v\n", o)
	}
}
