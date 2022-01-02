package main

import(
	"regexp"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

type StringCase struct {
	in string
	out string
}

func assert(t *testing.T, got string, want string, message string) {
	if got != want {
		t.Errorf("%s got: '%s', want: '%s'", message, got, want)
	}
}

func assertFun(t *testing.T, fun func(string) string, in string, want string) {
	got := fun(in)
	if got != want {
		t.Errorf("case '%s' got: '%s', want: '%s'", in, got, want)
	}
}

func asserts(t *testing.T, fun func(string) string, cases []StringCase) {
	for _, args := range cases {
		assertFun(t, fun, args.in, args.out)
	}
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
			<title>a title</title>
		</head>
		<body>
			<p>hi there</p>
			<p>world</p>
		</body>
		</html>
	`)
}

func TestGetDocumentMetaOutputs(t *testing.T) {
	gotMetaMessage, gotLongMeta := getDocumentMeta("host", getExampleHtml(t))
	assert(t, gotMetaMessage, "((host)) a title", "metaMessage")
	assert(t, gotLongMeta, "a title", "longMeta")
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
		})
}

func TestGetInputToTldrEdges(t *testing.T) {
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
		})
}
