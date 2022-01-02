package parser

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	lr "github.com/sirupsen/logrus"
	emoji "github.com/tmdvs/Go-Emoji-Utils"
	"golang.org/x/net/html"
)

func GetPreviews(url string, config OpenAI) (string, string) {
	resp, err := Client.Get(url)
	if err != nil {
		lr.Error(err)
	}
	contentType := resp.Header.Get("Content-Type")
	if !regexp.MustCompile("^text/html($|;)").MatchString(contentType) {
		return "", ""
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		lr.Error(err)
	}

	metaMessage, longMeta := getDocumentMeta(resp.Request.URL.Host, doc)
	tldrMessage := ""
	tldrInput := getInputToTldr(url, doc, longMeta)
	if len(tldrInput) > 0 {
		maybeTldr := trim(parseOpenAI(tldrInput, config))
		if len(maybeTldr) > 0 {
			tldrMessage = "((OpenAI TL;DR)) " + maybeTldr
		}
	}
	return metaMessage, tldrMessage
}

func trim(in string) string {
	trimmed := strings.Trim(in, " .,:\n")
	trimmed = regexp.MustCompile("[ \\t]+").ReplaceAllString(trimmed, " ")
	trimmed = regexp.MustCompile("\n+").ReplaceAllString(trimmed, "\n")
	return trimmed
}

func trimLastWord(in string) string {
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

func urlTokens(url string) string {
	tokens := url
	tokens = regexp.MustCompile("^[^:]+://").ReplaceAllString(tokens, "")
	tokens = regexp.MustCompile("[][ !\"#$%&'()*+,./:;<=>?@\\^_`{|}~-]+").ReplaceAllString(tokens, " ")
	tokens = regexp.MustCompile("(^| )www[0-9]*\\b *").ReplaceAllString(tokens, "")
	return tokens
}

func getTextWithSeparators(s *goquery.Selection, separator string) string {
	var buf bytes.Buffer
	for _, n := range s.Nodes {
		if n.Type == html.TextNode {
			buf.WriteString(n.Data)
			buf.WriteString(separator)
		}
	}

	return buf.String()
}

func getInputToTldr(pureUrl string, doc *goquery.Document, longMeta string) string {
	tldrInput := ""

	textNodes := doc.Find("h1").
		AddSelection(doc.Find("h2")).
		AddSelection(doc.Find("h3")).
		AddSelection(doc.Find("h4")).
		AddSelection(doc.Find("h5")).
		AddSelection(doc.Find("h6")).
		AddSelection(doc.Find("p")).
		AddSelection(doc.Find("summary")).
		AddSelection(doc.Find("li")).
		AddSelection(doc.Find("td")).
		AddSelection(doc.Find("button")).
		AddSelection(doc.Find("a")).
		AddSelection(doc.Find("div")).
		AddSelection(doc.Find("em")).
		AddSelection(doc.Find("strong")).
		AddSelection(doc.Find("i")).
		AddSelection(doc.Find("b")).
		AddSelection(doc.Find("pre")).
		AddSelection(doc.Find("code")).
		AddSelection(doc.Find("main")).
		AddSelection(doc.Find("article")).
		AddSelection(doc.Find("label")).
		AddSelection(doc.Find("span"))

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
