package main

import (
	"github.com/PuerkitoBio/goquery"
)

func getDocumentMeta(host string, doc *goquery.Document) (string, string) {
	metaMessage := ""
	longMeta := ""

	title := doc.Find("Title").Contents().Text()
	maybeLongMeta := title
	if len(maybeLongMeta) > 0 {
		longMeta = maybeLongMeta
		metaMessage = "((" + host + ")) " + longMeta
	}
	return metaMessage, longMeta
}

func getInputToTldr(pureUrl string, doc *goquery.Document, longMeta string) (string) {
	tldrInput := ""

	textNodes := doc.Find("p")

	pageContent := textNodes.Contents().Text()
	if len(pageContent) > 0 {
		tldrInput = pageContent
		tldrInput += "\ntl;dr:"
	}
	return tldrInput
}
