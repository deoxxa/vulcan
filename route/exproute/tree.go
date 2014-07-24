package exproute

import (
	"fmt"
	"github.com/mailgun/vulcan/location"
	"github.com/mailgun/vulcan/request"
	"strings"
)

type matchTree struct {
	root node
}

type node interface {
	canMerge(node) bool
	merge(node) (node, error)
	match(req request.Request) location.Location
}

// Leaf nodes always match the given location by any request.
type leaf struct {
	location location.Location
}

func (l *leaf) canMerge(node) bool {
	return false
}

func (l *leaf) merge(node) (node, error) {
	return nil, fmt.Errorf("Can't merge leaf node")
}

func (l *leaf) match(request.Request) location.Location {
	return l.location
}

type mapper interface {
	mapRequest(request.Request) string
}

type methodMapper struct {
}

func (m *methodMapper) mapRequest(r request.Request) string {
	return strings.ToUpper(r.GetHttpRequest().Method)
}

type switcher struct {
	cases []node
}

func newSwitcher() *switcher {
	return &switcher{
		cases: make([]node, 0),
	}
}

func (s *switcher) addCase(n node) {
	s.cases = append(s.cases, n)
}

func (s *switcher) canMerge(n node) bool {
	_, ok := n.(*switcher)
	return ok
}

func (s *switcher) merge(n node) (node, error) {
	other, ok := n.(*switcher)
	if !ok {
		return nil, fmt.Errorf("Can't merge %T to %T", s, n)
	}
	newS := &switcher{
		cases: make([]node, 0, len(s.cases)+len(other.cases)),
	}
	newS.cases = append(newS.cases, s.cases...)
	newS.cases = append(newS.cases, other.cases...)
	return newS, nil
}

func (s *switcher) match(r request.Request) location.Location {
	for _, c := range s.cases {
		loc := c.match(r)
		if loc != nil {
			return loc
		}
	}
	return nil
}

type predicateFn func(request.Request) bool

func makeMethodEq(method string) predicateFn {
	return func(r request.Request) bool {
		return strings.ToUpper(r.GetHttpRequest().Method) == strings.ToUpper(method)
	}
}

type predicate struct {
	child   node
	matchFn predicateFn
}

func (p *predicate) canMerge(node) bool {
	return false
}

func (p *predicate) merge(node) (node, error) {
	return nil, fmt.Errorf("Can't merge predicate node")
}

func (p *predicate) match(r request.Request) location.Location {
	if p.matchFn(r) {
		return p.child.match(r)
	}
	return nil
}
