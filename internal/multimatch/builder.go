package multimatch

// Builder allows for building of FSMs for constructing a Matcher.
// Safe for reuse after calling Build. Should not be copied by value.
type Builder struct {
	root  *node
	queue nodeQueue
}

func (b *Builder) Add(term []byte, output int) {
	current := b.ensureRoot()
	for _, b := range term {
		next := current.next[b]
		if next == nil {
			next = new(node)
			current.next[b] = next
		}
		current = next
	}
	current.appendUniqueOutput(output)
}

func (b *Builder) AddString(term string, output int) {
	b.Add([]byte(term), output)
}

// Build creates an FSM from the constructed trie
// and returns it as an immutable Matcher. Resets
// the current builder for reuse.
func (b *Builder) Build() Matcher {
	r := b.ensureRoot()
	b.buildFSMFromTrie()
	b.reset()
	return Matcher{root: r}
}

func (b *Builder) ensureRoot() *node {
	if b.root == nil {
		b.root = new(node)
	}
	return b.root
}

func (b Builder) buildFSMFromTrie() {
	b.root.failLink = b.root
	for _, child := range b.root.next {
		if child == nil {
			continue
		}
		child.failLink = b.root
		b.queue.push(child)
	}

	for !b.queue.empty() {
		node := b.queue.pop()
		for key, child := range node.next {
			if child == nil {
				continue
			}
			b.queue.push(child)
			fail := node.failLink
			for fail.next[key] == nil && fail != b.root {
				fail = fail.failLink
			}
			child.failLink = fail.next[key]
			if child.failLink == nil {
				child.failLink = b.root
			}
			for _, o := range child.failLink.output {
				child.appendUniqueOutput(o)
			}
		}
	}
}

func (b *Builder) reset() {
	b.root = nil
	b.queue.clear()
}
