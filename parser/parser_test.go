package parser

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

type StringCase struct {
	in  string
	out string
}

func showMiddleN(in string, cap int) string {
	out := in
	n := len(in)
	if n > cap {
		out = in[0:cap-cap/2-2] + "..." + in[n-cap/2+1:n] + " (" + strconv.Itoa(n) + ")"
	}
	return out
}

func showMiddle(in string) string {
	return showMiddleN(in, 64)
}

func assert(t *testing.T, got string, want string, message string) {
	if got != want {
		t.Errorf("%s got: '%s', want: '%s'", showMiddle(message), showMiddle(got), showMiddle(want))
	}
}

func assertFun(t *testing.T, fun func(string) string, in string, want string) {
	assert(t, fun(in), want, "case '"+in+"'")
}

func asserts(t *testing.T, fun func(string) string, cases []StringCase) {
	for _, args := range cases {
		assertFun(t, fun, args.in, args.out)
	}
}

func TestTrim(t *testing.T) {
	asserts(t, trim, []StringCase{
		{"a", "a"},
		{"a:b", "a:b"},
		{"a.,:.,:  ", "a"},
		{"  .,:a", "a"},
		{"a  b", "a b"},
		{"a\n\nb", "a\nb"},
	})
}

func TestTrimLastWord(t *testing.T) {
	asserts(t, trimLastWord, []StringCase{
		{"a", "a"},
		{"a  12345678901234567890123456789012", "a"},
		{"a  123456789012345678901234567890123", "a  123456789012345678901234567890123"},
	})
}

func TestUrlTokens(t *testing.T) {
	asserts(t, urlTokens, []StringCase{
		{"a", "a"},
		{"http://www.example.com/best-picture", "example com best picture"},
	})
}

func getHtmlFromString(t *testing.T, html string) *goquery.Document {
	reader := strings.NewReader(html)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		t.Errorf("Internal error: the HTML could not be parsed - %s", err)
	}
	return doc
}

func getExampleHtml(t *testing.T) *goquery.Document {
	return getHtmlFromString(t, `
		<html>
		<head>
			<title>a
bad  title</title>
		</head>
		<body>
			<p>hi there</p>
		</body>
		</html>
	`)
}

func testShowMiddleN(t *testing.T, n int, want string) {
	in := "123456789"
	got := showMiddleN(in, n)
	if got != want {
		t.Errorf("case ('%s',%d) got: '%s', want: '%s'", in, n, got, want)
	}
}

func TestShowMiddleN(t *testing.T) {
	testShowMiddleN(t, 9, "123456789")
	testShowMiddleN(t, 8, "12...789 (9)")
	testShowMiddleN(t, 7, "12...89 (9)")
}

func TestGetDocumentMetaOutputs(t *testing.T) {
	gotMetaMessage, gotLongMeta := getDocumentMeta("host", getExampleHtml(t))
	assert(t, gotMetaMessage, "((host)) a bad title", "metaMessage")
	assert(t, gotLongMeta, "a\nbad title", "longMeta")
}

func TestGetDocumentMetaLongTitle(t *testing.T) {
	cats := strings.Repeat("cat ", 96) + "cat"
	moreCats := cats + " cat cat"
	html := "<html><head><title>" + moreCats + "</title></head><body></body></html>"
	gotMetaMessage, gotLongMeta := getDocumentMeta("host", getHtmlFromString(t, html))
	assert(t, gotMetaMessage, "((host)) "+cats+"...", "metaMessage")
	assert(t, gotLongMeta, moreCats, "longMeta")
}

func TestGetDocumentMetaEdges(t *testing.T) {
	asserts(
		t,
		func(html string) string {
			fullHtml := "<html><head>" + html + "</head><body></body></html>"
			_, longMeta := getDocumentMeta("", getHtmlFromString(t, fullHtml))
			return longMeta
		},

		[]StringCase{
			{"<title>a b</title>", "a b"},
			{"<title> \n a  b \n </title>", "a b"},
		})
}

func TestGetInputToTldrOutput(t *testing.T) {
	gotTldrInput := getInputToTldr("http://example.com", getExampleHtml(t), "description")
	assert(t, gotTldrInput, "Keywords: example com.\ndescription\nhi there\ntl;dr:", "tldrInput")
}

func TestGetInputToTldrEmpty(t *testing.T) {
	gotTldrInput := getInputToTldr("http://example.com", getHtmlFromString(t, ""), "")
	assert(t, gotTldrInput, "", "tldrInput")
}

func TestGetInputToTldrEdges(t *testing.T) {
	enoughCats := strings.Repeat("cat ", 499) + "U"
	asserts(
		t,
		func(html string) string {
			fullHtml := "<html><body>" + html + "</body></html>"
			input := getInputToTldr("", getHtmlFromString(t, fullHtml), "")
			input = regexp.MustCompile("\ntl;dr:$").ReplaceAllString(input, "")
			return input
		},

		[]StringCase{
			{"<p>p</p>", "p"},
			{"<p>p1</p><p>p2</p>", "p1 p2"},
			{"<p>hi📞</p>", "hi"},
			{"<p>" + enoughCats + "</p>", enoughCats},
			{"<p>" + enoughCats + " cat</p>", enoughCats},
			{"<h1>h1</h1>", "h1"},
			{"<h1>h1</h1><p>p1</p><h2>h2</h2><p>p2</p>", "h1 h2 p1 p2"},
			{"<div><p>p</p> d <ul><li>li</li></ul> <script>script</script></div>", "p li d"},
		})
}
