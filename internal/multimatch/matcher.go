package multimatch

import (
	"iter"
)

type Matcher struct {
	root *node
}

func (m *Matcher) SearchString(term string) iter.Seq[int] {
	return m.Search([]byte(term))
}

func (m *Matcher) Search(term []byte) iter.Seq[int] {
	return func(yield func(int) bool) {
		var (
			pos   = 0
			state = m.root
		)
		for pos < len(term) {
			b := term[pos]
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
