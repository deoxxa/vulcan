package exproute

import (
	"fmt"
	"github.com/mailgun/vulcan/location"
	"github.com/mailgun/vulcan/request"
)

type trie struct {
	root *trieNode
}

func parseTrie(pattern string, matchNode node) (*trie, error) {
	t := &trie{
		root: &trieNode{},
	}
	if len(pattern) == 0 {
		return nil, fmt.Errorf("Empty pattern")
	}
	err := t.root.parsePattern(-1, []byte(pattern), matchNode)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (t *trie) canMerge(n node) bool {
	_, ok := n.(*trie)
	return ok
}

func (p *trie) merge(n node) (node, error) {
	other, ok := n.(*trie)
	if !ok {
		return nil, fmt.Errorf("Cant merge %T and %T")
	}
	root, err := p.root.merge(other.root)
	if err != nil {
		return nil, err
	}
	return &trie{root: root}, nil
}

func (p *trie) match(r request.Request) location.Location {
	if p.root == nil {
		return nil
	}

	path := r.GetHttpRequest().URL.Path
	if len(path) == 0 {
		path = "/"
	}
	return p.root.match(-1, []byte(path), r)
}

type trieNode struct {
	key       byte
	children  []*trieNode
	matchNode node
}

func (e *trieNode) isLeaf() bool {
	return e.matchNode != nil
}

func (e *trieNode) isRoot() bool {
	return e.key == byte(0)
}

func (e *trieNode) String() string {
	if e.isLeaf() {
		return fmt.Sprintf("leaf(%c)", e.key)
	} else if e.isRoot() {
		return fmt.Sprintf("root")
	} else {
		return fmt.Sprintf("node(%c)", e.key)
	}
}

func (e *trieNode) merge(o *trieNode) (*trieNode, error) {
	if e.key != o.key {
		return nil, fmt.Errorf("Can't merge nodes with different keys: %s and %s", e.key, o.key)
	}

	if e.isLeaf() && o.isLeaf() {
		return nil, fmt.Errorf("Can't merge two leaf nodes: %s and %s", e.String(), o.String())
	}

	if e.isLeaf() {
		return mergeWithLeaf(o, e)
	}

	if o.isLeaf() {
		return mergeWithLeaf(e, o)
	}

	children := make([]*trieNode, 0, len(e.children))
	merged := make(map[byte]bool)

	// First, find the nodes with similar keys and merge them
	for _, c := range e.children {
		for _, c2 := range o.children {
			if c.key == c2.key {
				m, err := c.merge(c2)
				if err != nil {
					return nil, err
				}
				merged[c.key] = true
				children = append(children, m)
			}
		}
	}

	// Next, append the keys that haven't been merged
	for _, c := range e.children {
		if !merged[c.key] {
			children = append(children, c)
		}
	}

	for _, c := range o.children {
		if !merged[c.key] {
			children = append(children, c)
		}
	}

	return &trieNode{key: e.key, children: children}, nil
}

func (p *trieNode) parsePattern(offset int, pattern []byte, matchNode node) error {
	// we are the root node or intermediate node
	if offset < len(pattern)-1 {
		p.children = []*trieNode{&trieNode{key: pattern[offset+1]}}
		//fmt.Printf("Pattern: '%s', I am '%s',  kid is '%s'\n", pattern, p.String(), p.children[0])
		return p.children[0].parsePattern(offset+1, pattern, matchNode)
	}
	// we are the leaf node
	//fmt.Printf("Pattern: %s, I am %s leaf\n", pattern, p)
	p.matchNode = matchNode
	return nil
}

func mergeWithLeaf(base *trieNode, leaf *trieNode) (*trieNode, error) {
	n := &trieNode{key: base.key, children: make([]*trieNode, len(base.children))}
	copy(n.children, base.children)
	n.matchNode = leaf.matchNode
	return n, nil
}

func (e *trieNode) match(offset int, path []byte, r request.Request) location.Location {
	// We are the root or the current key matches
	if offset == -1 || path[offset] == e.key {
		/*
			var o byte
			if offset >= 0 {
				o = path[offset]
			}
			fmt.Printf("Offset: %d, path[offset]: %s, e.key: %s, matchNode: %s\n", offset, o, e.key, e.matchNode)
		*/
		// This is a leaf node and we are at the last character of the pattern
		if e.matchNode != nil && offset == len(path)-1 {
			return e.matchNode.match(r)
		}
		// Check for the match in child nodes
		for _, c := range e.children {
			if loc := c.match(offset+1, path, r); loc != nil {
				return loc
			}
		}
	}
	return nil
}
