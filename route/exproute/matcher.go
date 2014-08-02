package exproute

import (
	"fmt"
	"github.com/mailgun/vulcan/location"
	"github.com/mailgun/vulcan/request"
	"regexp"
)

type matcher interface {
	canMerge(matcher) bool
	merge(matcher) (matcher, error)
	match(req request.Request) location.Location
}

type requestMapper func(req request.Request) string

func mapRequestToUrl(req request.Request) string {
	return req.GetHttpRequest().URL.String()
}

type regexpMatcher struct {
	mapper  requestMapper
	expr    *regexp.Regexp
	matcher matcher
}

func newRegexpMatcher(expr string, mapper requestMapper, matcher matcher) (matcher, error) {
	r, err := regexp.Compile(expr)

	if err != nil {
		return nil, fmt.Errorf("Bad regular expression: %s %s", expr, err)
	}
	return &regexpMatcher{expr: r, mapper: mapper, matcher: matcher}, nil
}

func (m *regexpMatcher) canMerge(matcher) bool {
	return false
}

func (m *regexpMatcher) merge(matcher) (matcher, error) {
	return nil, fmt.Errorf("Method not supported")
}

func (m *regexpMatcher) match(req request.Request) location.Location {
	if m.expr.MatchString(m.mapper(req)) {
		return m.matcher.match(req)
	}
	return nil
}

type constMatcher struct {
	location location.Location
}

func (c *constMatcher) canMerge(matcher) bool {
	return false
}

func (c *constMatcher) merge(matcher) (matcher, error) {
	return nil, fmt.Errorf("Method not supported")
}

func (c *constMatcher) match(req request.Request) location.Location {
	return c.location
}

type methodMatcher struct {
	methods []string
	matcher matcher
}

func (m *methodMatcher) canMerge(matcher) bool {
	return false
}

func (m *methodMatcher) merge(matcher) (matcher, error) {
	return nil, fmt.Errorf("Method not supported")
}

func (m *methodMatcher) match(req request.Request) location.Location {
	for _, c := range m.methods {
		if req.GetHttpRequest().Method == c {
			return m.matcher.match(req)
		}
	}
	return nil
}
