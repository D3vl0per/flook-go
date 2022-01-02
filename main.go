package main

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/D3vl0per/flook-go/parser"
	"github.com/PuerkitoBio/goquery"
	_ "github.com/joho/godotenv/autoload"
	lr "github.com/sirupsen/logrus"
	irc "github.com/thoj/go-ircevent"
	"golang.org/x/net/html"
	"mvdan.cc/xurls/v2"
)

type Config struct {
	IRC    IRC
	OpenAI parser.OpenAI
}

type IRC struct {
	Server  string
	Channel string
	Nick    string
	User    string
}

var config Config

func init() {
	lr.SetFormatter(&lr.TextFormatter{FullTimestamp: true})
	lr.SetOutput(os.Stdout)
	config.IRC.Server = os.Getenv("IRC_SERVER")
	config.IRC.Channel = os.Getenv("IRC_CHANNEL")
	config.IRC.Nick = os.Getenv("IRC_NICK")
	config.IRC.User = os.Getenv("IRC_USER")
	config.OpenAI.Instance = os.Getenv("OPENAI_INSTANCE")
	config.OpenAI.APIToken = os.Getenv("OPENAI_API_TOKEN")
	config.OpenAI.MaxToken, _ = strconv.Atoi(os.Getenv("OPENAI_MAX_TOKEN"))

}

/*
func getTitleHttpClient(url string) (title, host string){
	resp, err := client.Get(url)
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

func parseBody(url string) (resp *http.Response, body []byte) {
	resp = httpGet(url)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("[ERROR] %s", err.Error())
		}
	}()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		lr.Fatal(err)
	}
	return resp, body
}

func httpGet(url string) *http.Response {
	resp, err := parser.Client.Get(url)
	if err != nil {
		lr.Error(err)
	}

	return resp
}
func getRandomNitterInstance() (randomHost string) {
	resp, body := parseBody("https://raw.githubusercontent.com/xnaas/nitter-instances/master/history/summary.json")

	if resp.StatusCode == 200 {

		var nitterInstances NitterInstances
		if err := json.Unmarshal(body, &nitterInstances); err != nil {
			log.Printf("[ERROR] %s", err.Error())
		}
		var ups []int

		for i := range nitterInstances {
			if nitterInstances[i].Status == "up" && nitterInstances[i].TimeWeek < 1000 {
				ups = append(ups, i)
			}
		}

		rand.Seed(time.Now().UnixNano())
		randomHost = nitterInstances[ups[rand.Intn(len(ups)-1)]].Name
	} else {
		randomHost = "nitter.hu"
	}

	return randomHost
}

func main() {
	irccon := irc.IRC(config.IRC.Nick, config.IRC.User)
	if irccon == nil {
		log.Fatal("IRC connection failed")
	}
	irccon.VerboseCallbackHandler = false
	irccon.Debug = false
	irccon.UseTLS = true
	irccon.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	irccon.AddCallback("001", func(e *irc.Event) { irccon.Join(config.IRC.Channel) })
	//irccon.AddCallback("366", func(e *irc.Event) {  })
	irccon.AddCallback("PRIVMSG", func(event *irc.Event) {

		rxRelaxed := xurls.Relaxed()
		re := regexp.MustCompile(" [(]re:? @[^ :]+: .*")
		message := re.ReplaceAllString(event.Message(), "")
		urlsInMessage := rxRelaxed.FindAllString(message, -1)
		//irccon.Privmsg(event.Arguments[0], "Lol, twitter")
		if len(urlsInMessage) > 0 {
			if strings.Contains(urlsInMessage[0], "twitter.com/") {

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
					irccon.Privmsg(event.Arguments[0], "((Content)) "+tweetContent)
				}

			} else if strings.Contains(urlsInMessage[0], "reddit.com/") {

				pureUrl := urlsInMessage[0]
				replacedUrl := strings.Replace(pureUrl, "reddit.com", "libreddit.hu", -1)
				irccon.Privmsg(event.Arguments[0], replacedUrl)

			} else {
				meta, tldr := parser.GetPreviews(urlsInMessage[0], config.OpenAI)
				if len(meta) > 0 {
					irccon.Privmsg(event.Arguments[0], meta)
				}
				if len(tldr) > 0 {
					irccon.Privmsg(event.Arguments[0], tldr)
				}

			}
		}
	})
	err := irccon.Connect(config.IRC.Server)
	if err != nil {
		lr.Error(err)
		return
	}
	irccon.Loop()
}
