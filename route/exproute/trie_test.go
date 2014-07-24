package exproute

import (
	"bytes"
	"fmt"
	"github.com/mailgun/vulcan/location"
	"github.com/mailgun/vulcan/netutils"
	"github.com/mailgun/vulcan/request"
	"github.com/mailgun/vulcan/testutils"
	. "gopkg.in/check.v1"
	"net/http"
	"strings"
	"testing"
)

func TestTrie(t *testing.T) { TestingT(t) }

type TrieSuite struct {
}

var _ = Suite(&TrieSuite{})

func (s *TrieSuite) TestParseTrieSuccess(c *C) {
	t, l := makeTrie(c, "/", makeLoc("loc1"))
	c.Assert(t.match(makeReq("http://google.com")), Equals, l.location)
}

func (s *TrieSuite) TestPrintTrie(c *C) {
	t, _ := makeTrie(c, "/a", makeLoc("loc1"))
	expected := `
root
 node(/)
  leaf(a)
`
	c.Assert(printTrie(t), Equals, expected)
}

func (s *TrieSuite) TestMergeTriesCommonPrefix(c *C) {
	t1, l1 := makeTrie(c, "/a", makeLoc("loc1"))
	t2, l2 := makeTrie(c, "/b", makeLoc("loc2"))

	t3, err := t1.merge(t2)
	c.Assert(err, IsNil)

	expected := `
root
 node(/)
  leaf(a)
  leaf(b)
`
	c.Assert(printTrie(t3.(*trie)), Equals, expected)

	c.Assert(t3.match(makeReq("http://google.com/a")), Equals, l1.location)
	c.Assert(t3.match(makeReq("http://google.com/b")), Equals, l2.location)
}

func (s *TrieSuite) TestMergeTriesSubtree(c *C) {
	t1, l1 := makeTrie(c, "/aa", makeLoc("loc1"))
	t2, l2 := makeTrie(c, "/a", makeLoc("loc2"))

	t3, err := t1.merge(t2)
	c.Assert(err, IsNil)

	expected := `
root
 node(/)
  leaf(a)
   leaf(a)
`
	c.Assert(printTrie(t3.(*trie)), Equals, expected)

	c.Assert(t3.match(makeReq("http://google.com/aa")), Equals, l1.location)
	c.Assert(t3.match(makeReq("http://google.com/a")), Equals, l2.location)
}

func (s *TrieSuite) TestMergeCases(c *C) {
	testCases := []struct {
		trees    []string
		url      string
		expected string
	}{
		{
			[]string{"/v2/domains/", "/v2/domains/domain1"},
			"http://google.com/v2/domains/domain1",
			"/v2/domains/domain1",
		},
	}
	for _, tc := range testCases {
		t, _ := makeTrie(c, tc.trees[0], makeLoc(tc.trees[0]))
		for i, pattern := range tc.trees {
			if i == 0 {
				continue
			}
			t2, _ := makeTrie(c, pattern, makeLoc(pattern))
			out, err := t.merge(t2)
			c.Assert(err, IsNil)
			t = out.(*trie)
		}
		out := t.match(makeReq(tc.url))
		c.Assert(out.(*location.ConstHttpLocation).Url, Equals, tc.expected)
	}
}

func (s *TrieSuite) BenchmarkMatching(c *C) {
	rndString := testutils.NewRndString()
	l := makeLoc("loc")

	t, _ := makeTrie(c, rndString.MakePath(20, 10), l)

	for i := 0; i < 10000; i++ {
		t2, _ := makeTrie(c, rndString.MakePath(20, 10), l)
		out, err := t.merge(t2)
		if err != nil {
			c.Assert(err, IsNil)
		}
		t = out.(*trie)
	}
	req := makeReq(fmt.Sprintf("http://google.com/%s", rndString.MakePath(20, 10)))

	for i := 0; i < c.N; i++ {
		t.match(req)
	}
}

func makeTrie(c *C, path string, location location.Location) (*trie, *leaf) {
	l := &leaf{
		location: location,
	}
	t, err := parseTrie(path, l)
	c.Assert(err, IsNil)
	c.Assert(t, NotNil)
	return t, l
}

func makeReq(url string) request.Request {
	u := netutils.MustParseUrl(url)
	return &request.BaseRequest{
		HttpRequest: &http.Request{URL: u},
	}
}

func makeLoc(url string) location.Location {
	return &location.ConstHttpLocation{Url: url}
}

func printTrie(t *trie) string {
	return printTrieNode(t.root)
}

func printTrieNode(e *trieNode) string {
	out := &bytes.Buffer{}
	printTrieNodeInner(out, e, 0)
	return out.String()
}

func printTrieNodeInner(b *bytes.Buffer, e *trieNode, offset int) {
	padding := strings.Repeat(" ", offset)
	if e.isLeaf() {
		fmt.Fprintf(b, "%sleaf(%c)\n", padding, rune(e.key))
	} else if e.isRoot() {
		fmt.Fprintf(b, "\nroot\n")
	} else {
		fmt.Fprintf(b, "%snode(%c)\n", padding, rune(e.key))
	}
	if len(e.children) != 0 {
		for _, c := range e.children {
			printTrieNodeInner(b, c, offset+1)
		}
	}
}
