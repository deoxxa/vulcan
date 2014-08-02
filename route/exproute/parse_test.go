package exproute

import (
	. "gopkg.in/check.v1"
)

type ParseSuite struct {
}

var _ = Suite(&ParseSuite{})

func (s *TrieSuite) TestParseSuccess(c *C) {
	testCases := []struct {
		Expression string
		Url        string
		Method     string
	}{
		{
			`TrieRoute("/helloworld")`,
			`http://google.com/helloworld`,
			"GET",
		},
		{
			`TrieRoute("GET", "/helloworld")`,
			`http://google.com/helloworld`,
			"GET",
		},
		{
			`TrieRoute("GET", "POST", "/helloworld")`,
			`http://google.com/helloworld`,
			"POST",
		},
		{
			`TrieRoute("/hello/<world>")`,
			`http://google.com/hello/world`,
			"GET",
		},
		{
			`RegexpRoute("/helloworld")`,
			`http://google.com/helloworld`,
			"GET",
		},
		{
			`RegexpRoute("POST", "/helloworld")`,
			`http://google.com/helloworld`,
			"POST",
		},
	}
	for _, tc := range testCases {
		l := makeLoc(tc.Url)
		m, err := parseExpression(tc.Expression, l)
		c.Assert(err, IsNil)
		c.Assert(m, NotNil)

		req := makeReq(tc.Url)
		req.GetHttpRequest().Method = tc.Method
		outLoc := m.match(req)
		c.Assert(outLoc, Equals, l)
	}
}

func (s *TrieSuite) TestParseFailures(c *C) {
	testCases := []string{
		`bad`,                                       // unsupported identifier
		`bad expression`,                            // not a valid go expression
		`TrieRoute("/path") || TrieRoute("/path2")`, // complex logic
		`1 && 2`,                          // unsupported statements
		`"standalone literal"`,            // standalone literal
		`UnknownFunction("hi")`,           // unknown functin
		`TrieRoute(1)`,                    // bad argument type
		`RegexpRoute(1)`,                  // bad argument type
		`TrieRoute()`,                     // no arguments
		`RegexpRoute()`,                   // no arguments
		`TrieRoute(RegexpRoute("hello"))`, // nested calls
		`TrieRoute("")`,                   // bad trie expression
		`RegexpRoute("[[[[")`,             // bad regular expression
	}

	for _, expr := range testCases {
		m, err := parseExpression(expr, makeLoc("loc1"))
		c.Assert(err, NotNil)
		c.Assert(m, IsNil)
	}
}
