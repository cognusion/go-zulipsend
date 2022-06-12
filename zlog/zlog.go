package main

import (
	zulip "github.com/cognusion/go-zulipsend"

	"flag"
	"log"
	"os"
)

var (
	baseURL string
	user    string
	token   string
	stream  string
	topic   string
	message string
	debug   bool
)

func main() {
	flag.StringVar(&baseURL, "url", "", "Base URL to Zulip host")
	flag.StringVar(&user, "user", "", "User to authenticate as")
	flag.StringVar(&token, "token", "", "Token to authenticate with")
	flag.StringVar(&stream, "stream", "", "Stream to use")
	flag.StringVar(&topic, "topic", "", "Topic to post to")
	flag.StringVar(&message, "message", "", "Message to send")
	flag.BoolVar(&debug, "debug", false, "Enable debug output")

	flag.Parse()

	if debug {
		zulip.DebugOut = log.New(os.Stderr, "[DEBUG]", log.Lshortfile)
	}

	z := zulip.Zulip{BaseURL: baseURL, Username: user, Token: token}
	zw := z.ToWriter(stream, topic)

	ZulipOut := log.New(zw, "", log.LUTC)

	ZulipOut.Println(message)

}
