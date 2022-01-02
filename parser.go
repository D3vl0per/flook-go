package main

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/tmdvs/Go-Emoji-Utils"
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
		tldrInput = emoji.RemoveAll(tldrInput)
		if len(tldrInput) > 2000 {
			tldrInput = tldrInput[0:2000]
		}
		tldrInput += "\ntl;dr:"
	}
	return tldrInput
}
