package main

import (
	zulip "github.com/cognusion/go-zulipsend"

	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

var (
	baseURL       string
	user          string
	token         string
	stream        string
	topic         string
	message       string
	retryCount    int
	retryInterval time.Duration
	debug         bool
)

func main() {
	flag.StringVar(&baseURL, "url", "", "Base URL to Zulip host")
	flag.StringVar(&user, "user", "", "User to authenticate as")
	flag.StringVar(&token, "token", "", "Token to authenticate with")
	flag.StringVar(&stream, "stream", "", "Stream to use")
	flag.StringVar(&topic, "topic", "", "Topic to post to")
	flag.StringVar(&message, "message", "", "Message to send")
	flag.IntVar(&retryCount, "retries", 0, "Number of retries on send fail (0 disables retries)")
	flag.DurationVar(&retryInterval, "interval", 1*time.Second, "Interval to retry")
	flag.BoolVar(&debug, "debug", false, "Enable debug output")

	flag.Parse()

	if debug {
		zulip.DebugOut = log.New(os.Stderr, "[DEBUG]", log.Lshortfile)
	}

	z := zulip.Zulip{
		BaseURL:  baseURL,
		Username: user,
		Token:    token,
		Retries:  retryCount,
		Interval: retryInterval,
	}

	err := z.Send(stream, topic, message)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
