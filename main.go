package main

import (
	"crypto/tls"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"encoding/json"
	_ "github.com/joho/godotenv/autoload"
	lr "github.com/sirupsen/logrus"
	"github.com/thoj/go-ircevent"
	"golang.org/x/net/html"
	"mvdan.cc/xurls/v2"
	"math/rand"

	"github.com/PuerkitoBio/goquery"
	"github.com/tmdvs/Go-Emoji-Utils"
	"strconv"
	"bytes"
)

type Config struct {
	IRC IRC
	OpenAI OpenAI
}

type IRC struct {
	Server string
	Channel string
	Nick string
	User string
}

type OpenAI struct {
	Instance string
	APIToken string
	MaxToken int
}

var config Config

func init() {
	lr.SetFormatter(&lr.TextFormatter{FullTimestamp: true})
	lr.SetOutput(os.Stdout)
	config.IRC.Server = os.Getenv("IRC_SERVER")
	config.IRC.Channel = os.Getenv("IRC_CHANNEL")
	config.IRC.Nick = os.Getenv("IRC_NICK")
	config.IRC.User	= os.Getenv("IRC_USER")
	config.OpenAI.Instance = os.Getenv("OPENAI_INSTANCE")
	config.OpenAI.APIToken = os.Getenv("OPENAI_API_TOKEN")
	config.OpenAI.MaxToken, _ = strconv.Atoi(os.Getenv("OPENAI_MAX_TOKEN"))

}
/*
func getTitleHttpClient(url string) (title, host string){
	c := &http.Client{
		Timeout: time.Duration(5 * time.Second),
		Transport: &http.Transport{
				MaxIdleConns:       10,
				IdleConnTimeout:    30 * time.Second,
				DisableCompression: false,
		},
	}

	resp, err := c.Get(url)
    if err != nil {
        lr.Error(err)
    }

	defer resp.Body.Close()
	host = resp.Request.URL.Host
	if title, ok := GetHtmlTitle(resp.Body); ok {
		return title, host
	} else {
		return "", host
	}
	
}
*/
func isTitleElement(n *html.Node) bool {
	return n.Type == html.ElementNode && n.Data == "title"
}

func traverse(n *html.Node) (string, bool) {
	if isTitleElement(n) {
		return n.FirstChild.Data, true
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result, ok := traverse(c)
		if ok {
			return result, ok
		}
	}

	return "", false
}

func GetHtmlTitle(r io.Reader) (string, bool) {
	doc, err := html.Parse(r)
	if err != nil {
		panic("Fail to parse html")
	}

	return traverse(doc)
}

type NitterInstances []struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Status   string `json:"status"`
	TimeWeek int    `json:"timeWeek"`
}

func parseBody(url string) (resp *http.Response, body []byte){
	resp = httpGet(url)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
			lr.Fatal(err)
	}
	return resp, body
}

func httpGet(url string) (*http.Response) {

	c := &http.Client{
		Timeout: time.Duration(5 * time.Second),
		Transport: &http.Transport{
				MaxIdleConns:       10,
				IdleConnTimeout:    30 * time.Second,
				DisableCompression: false,
		},
	}

	resp, err := c.Get(url)
    if err != nil {
        lr.Error(err)
    }

	return resp
}
func getRandomNitterInstance() (randomHost string){
	resp, body := parseBody("https://raw.githubusercontent.com/xnaas/nitter-instances/master/history/summary.json")

	if resp.StatusCode == 200 {
	
		var nitterInstances NitterInstances
		json.Unmarshal(body, &nitterInstances)
		var ups []int
		
		for i := range nitterInstances {
			if nitterInstances[i].Status == "up" && nitterInstances[i].TimeWeek < 1000 {
				ups = append(ups, i)
			}
		}

		rand.Seed(time.Now().UnixNano())
		randomHost = nitterInstances[ups[rand.Intn(len(ups) - 1)]].Name
	} else {
		randomHost = "nitter.hu"
	}

	return randomHost
}

type Payload struct {
	Prompt           string  `json:"prompt"`
	Temperature      float64 `json:"temperature"`
	MaxTokens        int     `json:"max_tokens"`
	TopP             float64 `json:"top_p"`
	FrequencyPenalty float64 `json:"frequency_penalty"`
	PresencePenalty  float64 `json:"presence_penalty"`
}

type OpenAIResponse struct {
	ID      string    `json:"id"`
	Object  string    `json:"object"`
	Created int       `json:"created"`
	Model   string    `json:"model"`
	Choices []Choices `json:"choices"`
}
type Choices struct {
	Text         string      `json:"text"`
	Index        int         `json:"index"`
	Logprobs     interface{} `json:"logprobs"`
	FinishReason string      `json:"finish_reason"`
}

func parseOpenAI(source string) (tldr string){
	resp := openAIHttpPost(source)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
			lr.Fatal(err)
	}


	var openAIResponse OpenAIResponse
	json.Unmarshal(body, &openAIResponse)

	return openAIResponse.Choices[0].Text
}

func openAIHttpPost(source string) (tldr *http.Response){

	if len(source) > 2000 {
		source = source[0:2000]
	}
	
	source = emoji.RemoveAll(source)
	source += "\ntl;dr:"

	data := Payload{
		Prompt: source,
		Temperature: 0.3,
		MaxTokens: config.OpenAI.MaxToken,
		TopP: 1.0,
		FrequencyPenalty: 0.0,
		PresencePenalty: 0.0,
	}
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		lr.Error(err)
	}
	body := bytes.NewReader(payloadBytes)
	
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/engines/" + config.OpenAI.Instance + "/completions", body)
	if err != nil {
		lr.Error(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", config.OpenAI.APIToken)
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		lr.Error(err)
	}
	return resp

}

func main() {
		
        irccon := irc.IRC(config.IRC.Nick, config.IRC.User)
        irccon.VerboseCallbackHandler = false
        irccon.Debug = false
        irccon.UseTLS = true
        irccon.TLSConfig = &tls.Config{InsecureSkipVerify: true}
        irccon.AddCallback("001", func(e *irc.Event) { irccon.Join(config.IRC.Channel) })
        //irccon.AddCallback("366", func(e *irc.Event) {  })
		irccon.AddCallback("PRIVMSG", func(event *irc.Event) {

			rxRelaxed := xurls.Relaxed()
			urlsInMessage := rxRelaxed.FindAllString(event.Message(), -1)
			//irccon.Privmsg(event.Arguments[0], "Lol, twitter")
			if len(urlsInMessage) > 0 {
				if (strings.Contains(urlsInMessage[0], "twitter.com/")){
					
					

					pureUrl := urlsInMessage[0]
					
					replacedUrl := strings.Replace(pureUrl, "twitter.com", getRandomNitterInstance(), -1)
					irccon.Privmsg(event.Arguments[0], replacedUrl)

					resp := httpGet(replacedUrl)
					doc, err := goquery.NewDocumentFromReader(resp.Body)
					if err != nil {
						lr.Error(err)
					}
					
					tweetContent := doc.Find(".main-tweet").Find(".tweet-content").Contents().Text()
					//title := doc.Find("Title").Contents().Text()
					if len(tweetContent) > 0 {
						irccon.Privmsg(event.Arguments[0], "((Content)) " + tweetContent)
					}
					

				} else if (strings.Contains(urlsInMessage[0], "reddit.com/")){

					pureUrl := urlsInMessage[0]
					replacedUrl := strings.Replace(pureUrl, "reddit.com", "libreddit.hu", -1)
					irccon.Privmsg(event.Arguments[0], replacedUrl)

				} else {
					pureUrl := urlsInMessage[0]

					resp := httpGet(pureUrl)
					doc, err := goquery.NewDocumentFromReader(resp.Body)
					if err != nil {
						lr.Error(err)
					}
					title := doc.Find("Title").Contents().Text()
					if len(title) > 0 {
						irccon.Privmsg(event.Arguments[0], "((" + resp.Request.URL.Host + ")) " + title)
					}

					tldr := parseOpenAI(doc.Find("p").Contents().Text())

					if len(tldr) > 0 {
						irccon.Privmsg(event.Arguments[0], "((Estimated TL;DR)) " + tldr)
					}

				}	
			}
		});
        err := irccon.Connect(config.IRC.Server)
	if err != nil {
		lr.Error(err)
		return
	}
        irccon.Loop()
}

