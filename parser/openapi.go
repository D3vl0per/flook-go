package parser

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	lr "github.com/sirupsen/logrus"
)

func parseOpenAI(source string, config OpenAI) (tldr string) {
	resp := openAIHttpPost(source, config)

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("[ERROR] %s", err.Error())
		}
	}()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		lr.Fatal(err)
	}

	var openAIResponse OpenAIResponse
	if err := json.Unmarshal(body, &openAIResponse); err != nil {
		log.Printf("[ERROR] %s", err.Error())
	}

	return openAIResponse.Choices[0].Text
}

func openAIHttpPost(source string, config OpenAI) (tldr *http.Response) {
	data := Payload{
		Prompt:           source,
		Temperature:      0.3,
		MaxTokens:        config.MaxToken,
		TopP:             1.0,
		FrequencyPenalty: 0.0,
		PresencePenalty:  0.0,
	}
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		lr.Error(err)
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/engines/"+config.Instance+"/completions", body)
	if err != nil {
		lr.Error(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", config.APIToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		lr.Error(err)
	}
	return resp
}
