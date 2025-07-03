package multimatch

import (
	"iter"
)

type Matcher struct {
	root *node
}

func (m *Matcher) SearchString(text string) iter.Seq[int] {
	return m.Search([]byte(text))
}

func (m *Matcher) Search(text []byte) iter.Seq[int] {
	return func(yield func(int) bool) {
		var (
			pos   = 0
			state = m.root
		)
		for pos < len(text) {
			b := text[pos]
			if state.next[b] != nil {
				state = state.next[b]
				for _, o := range state.output {
					if !yield(o) {
						return
					}
				}
				pos++
			} else if state == m.root {
				pos++
			} else {
				state = state.failLink
			}
		}
	}
}
