package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type pr struct {
	Action      string
	Number      int
	PullRequest struct {
		Head struct {
			SHA string
		}
	} `json:"pull_request"`
	Repository struct {
		FullName    string `json:"full_name"`
		StatusesURL string `json:"statuses_url"` // set in events, contains {sha} placeholder
	}
	StatusesURL string `json:"statuses_url"` // set when getting manually
}

type prState string

const (
	statePending prState = "pending"
	stateSuccess         = "success"
	stateError           = "error"
	stateFailure         = "failure"
)

func (p *pr) setStatus(state prState, context, description, username, token string) {
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(map[string]string{
		"state":       string(state),
		"description": description,
		"context":     context,
	})

	url := p.StatusesURL
	if url == "" {
		url = p.Repository.StatusesURL
	}
	url = strings.Replace(url, "{sha}", p.PullRequest.Head.SHA, 1)

	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		log.Println("Request:", err)
		return
	}
	req.SetBasicAuth(username, token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Post:", err)
		return
	}
	resp.Body.Close()

	if resp.StatusCode > 299 {
		log.Println("Post:", resp.Status)
		return
	}

}