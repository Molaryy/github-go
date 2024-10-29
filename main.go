package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	port = "8082"
)

type Push struct {
	user          string
	userURL       string
	commitMessage string
	commitUrl     string
	branchName    string
	time          string
}

type Issue struct {
	action      string
	url         string
	title       string
	description string
	author      string
	authorUrl   string
	createdTime string
}

type Branch struct {
	action           string
	name             string
	author           string
	authorUrl        string
	masterBranchName string
	description      string
	url              string
	createdTime      string
}

func getPush(p *Push) string {
	return fmt.Sprintf("> ## `🚀 Push`\\n> \\n> **👤 User**: [`%s`](%s)\\n> **✉️ Commit**: [`%s`](%s)\\n> ** 🎋 Branch**:  `%s`\\n> **📅 Date**: `%s`",
		p.user, p.userURL, p.commitMessage, p.commitUrl, p.branchName, p.time)
}

func getPushSlack(p *Push) string {
	return fmt.Sprintf("> `🚀 Push`\\n> \\n> *👤 User*: [`%s`](%s)\\n> *✉️ Commit*: [`%s`](%s)\\n> * 🎋 Branch*:  `%s`\\n> *📅 Date*: `%s`",
		p.user, p.userURL, p.commitMessage, p.commitUrl, p.branchName, p.time)
}

func getIssue(i *Issue) string {
	return fmt.Sprintf("> ## `Issue %s`\\n> \\n> **Created by**: [`%s`](%s)\\n> **Title**: `%s`\\n> **Description**:  `%s`\\n> **Created at**: `%s`",
		i.action, i.author, i.authorUrl, i.title, i.description, i.createdTime)
}

func getBranch(b *Branch) string {
	return fmt.Sprintf("> ## `Branch %s`\\n> \\n> **Created by**: [`%s`](%s)\\n> **Name**: `%s`\\n> **Description**:  `%s`\\n> **Created at**: `%s`",
		b.action, b.author, b.authorUrl, b.name, b.description, b.createdTime)
}

func sendToSlack(channelID string, text string) {
    botToken := "Bearer " + os.Getenv("SLACK_BOT_TOKEN")
    url := "https://slack.com/api/chat.postMessage"
    method := "POST"
    
    b := fmt.Sprintf("{\"channel\": \"%s\",\"text\": \"%s\"}", channelID, text)
    payload := strings.NewReader(b)
    client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", botToken)
	res, err := client.Do(req)
	defer res.Body.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
}

func sendToDiscord(content string) {
	botToken := "Bot " + os.Getenv("DISCORD_BOT_TOKEN")
	url := "https://discord.com/api/channels/1291527142902993047/messages"
	method := "POST"

	b := fmt.Sprintf("{\"content\": \"%s\"}", content)
	fmt.Println(b)
	payload := strings.NewReader(b)
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", botToken)

	res, err := client.Do(req)
	defer res.Body.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
}

func createPush(body string) *Push {
	var result map[string]any
	json.Unmarshal([]byte(body), &result)

	headCommit := result["head_commit"].(map[string]any)
	sender := result["sender"].(map[string]any)
	userURL := sender["html_url"].(string)
	message := headCommit["message"].(string)
	timestamp := headCommit["timestamp"]
	author := headCommit["author"].(map[string]any)
	commitURL := headCommit["url"].(string)
	user := author["username"].(string)
	ref := result["ref"].(string)
	ref = strings.Replace(ref, "refs/heads/", "", -1)

	t, _ := time.Parse(time.RFC3339, timestamp.(string))
	formattedTime := t.Format("2006-01-02 15:04")

	return &Push{
		user:          user,
		userURL:       userURL,
		commitMessage: message,
		branchName:    ref,
		commitUrl:     commitURL,
		time:          formattedTime,
	}
}

func jsonToMap(jsonStr string) map[string]interface{} {
	result := make(map[string]interface{})
	json.Unmarshal([]byte(jsonStr), &result)
	return result
}

func h(req *http.Request) {
    // Access all headers
    for name, values := range req.Header {
        // Iterate over all values for the header key
        for _, value := range values {
            fmt.Printf("%s: %s\n", name, value)
        }
    }
}

func handlerGit(w http.ResponseWriter, req *http.Request) {
	action := req.Header.Get("X-GitHub-Event")
	b, err := io.ReadAll(req.Body)
	if err != nil {
		// TODO: coulnd't read body
	}
	h(req)

	body := string(b[:])
	if action == "" {
		// TODO: log here
	}
	switch action {
	case "push":
		sendToDiscord(getPush(createPush(body)))
		sendToSlack("C07S8SZKHT9", getPushSlack(createPush(body)))
	}
}

func handlerSlack(w http.ResponseWriter, req *http.Request) {
	b, err := io.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}

	body := string(b[:])

	bodyMap := jsonToMap(body)
	event := bodyMap["event"].((map[string]any))
	text := event["text"].(string)
	fmt.Println(text)
	sendToDiscord(text)
	fmt.Fprint(w, text)
}

func main() {

	fmt.Println("Listening on port " + port)
	http.HandleFunc("/github/event", handlerGit)
	http.HandleFunc("/slack", handlerSlack)
	http.ListenAndServe(":"+port, nil)
}
