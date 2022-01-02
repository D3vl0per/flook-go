package main

import (
	"bytes"
	"strings"
	"regexp"

	"github.com/PuerkitoBio/goquery"
	"github.com/tmdvs/Go-Emoji-Utils"
	"golang.org/x/net/html"
)

func trim(in string) (string) {
	trimmed := strings.Trim(in, " .,:\n")
	trimmed = regexp.MustCompile("[ \\t]+").ReplaceAllString(trimmed, " ")
	trimmed = regexp.MustCompile("\n+").ReplaceAllString(trimmed, "\n")
	return trimmed
}

func trimLastWord(in string) (string) {
	return regexp.MustCompile("\\s+\\S{0,32}$").ReplaceAllString(in, "")
}

func getDocumentMeta(host string, doc *goquery.Document) (string, string) {
	metaMessage := ""
	longMeta := ""

	title := doc.Find("Title").Contents().Text()
	maybeLongMeta := trim(title)
	if len(maybeLongMeta) > 0 {
		longMeta = maybeLongMeta
		onelineLongMeta := regexp.MustCompile("\\s+").ReplaceAllString(longMeta, " ")
		metaMessage = "((" + host + ")) " + onelineLongMeta
		if len(metaMessage) > 397 {
			metaMessage = trimLastWord(metaMessage[0:397]) + "..."
		}
	}
	return metaMessage, longMeta
}


func urlTokens(url string) (string) {
	tokens := url
	tokens = regexp.MustCompile("^[^:]+://").ReplaceAllString(tokens, "")
	tokens = regexp.MustCompile("[][ !\"#$%&'()*+,./:;<=>?@\\^_`{|}~-]+").ReplaceAllString(tokens, " ")
	tokens = regexp.MustCompile("(^| )www[0-9]*\\b *").ReplaceAllString(tokens, "")
	return tokens
}

func getTextWithSeparators(s *goquery.Selection, separator string) string {
	var buf bytes.Buffer

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			buf.WriteString(n.Data)
			buf.WriteString(separator)
		}
		if n.FirstChild != nil {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c)
			}
		}
	}
	for _, n := range s.Nodes {
		f(n)
	}

	return buf.String()
}

func getInputToTldr(pureUrl string, doc *goquery.Document, longMeta string) (string) {
	tldrInput := ""

	textNodes := doc.Find("p")
	pageContent := trim(getTextWithSeparators(textNodes.Contents(), "\n"))

	pageContent = emoji.RemoveAll(pageContent)
	if len(pageContent) > 0 {
		url := urlTokens(pureUrl)
		if len(url) > 0 {
			url = "Keywords: " + emoji.RemoveAll(urlTokens(pureUrl)) + ".\n"
		}
		if len(longMeta) > 0 {
			longMeta = emoji.RemoveAll(longMeta) + "\n"
		}
		tldrInput = url + longMeta + pageContent

		if len(tldrInput) > 2000 {
			tldrInput = trimLastWord(tldrInput[0:2000])
		}
		tldrInput += "\ntl;dr:"
	}
	return tldrInput
}
