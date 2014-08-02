package exproute

import (
	. "gopkg.in/check.v1"
)

type ParseSuite struct {
}

var _ = Suite(&ParseSuite{})

func (s *TrieSuite) TestParseSuccess(c *C) {
	m, err := parseExpression(`TrieRoute("/helloworld")`, nil)
	c.Assert(err, IsNil)
	c.Assert(m, NotNil)
}
