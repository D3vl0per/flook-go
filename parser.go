package main

import (
	"strings"
	"regexp"

	"github.com/PuerkitoBio/goquery"
	"github.com/tmdvs/Go-Emoji-Utils"
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

func getInputToTldr(pureUrl string, doc *goquery.Document, longMeta string) (string) {
	tldrInput := ""

	textNodes := doc.Find("p")

	pageContent := trim(textNodes.Contents().Text())
	if len(pageContent) > 0 {
		tldrInput = pageContent
		tldrInput = emoji.RemoveAll(tldrInput)
		if len(tldrInput) > 2000 {
			tldrInput = trimLastWord(tldrInput[0:2000])
		}
		tldrInput += "\ntl;dr:"
	}
	return tldrInput
}
