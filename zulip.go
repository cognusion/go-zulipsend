package zulip

import (
	"github.com/eapache/go-resiliency/retrier"

	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	// DebugOut is a log.Logger that gets debug output. Hook it up to get output
	DebugOut = log.New(ioutil.Discard, "[DEBUG]", log.Lshortfile)
)

// Zulip is a struct to enable repeatable transmissions to a Zulip instance
type Zulip struct {
	BaseURL  string
	Username string
	Token    string
	Retries  int           // Number of times to retry sending a message (disabled with 0)
	Interval time.Duration // Interval between retries
}

// ToWriter returns an io.Writer (zulip.Writer) suitable of being pumped into a log.New or anywhere
// else you can use an io.Writer
func (z *Zulip) ToWriter(stream, topic string) io.Writer {
	zw := writer{z, stream, topic}
	return &zw
}

// Send a message to Zulip, possibly retrying if Interval > 0
func (z *Zulip) Send(stream, topic, message string) (err error) {

	pBody := []string{
		fmt.Sprintf("type=%s", "stream"),
		fmt.Sprintf("to=%s", url.QueryEscape(stream)),
		fmt.Sprintf("subject=%s", url.QueryEscape(topic)),
		fmt.Sprintf("content=%s", url.QueryEscape(message)),
	}

	dencode := strings.Join(pBody, "&")
	DebugOut.Printf("Zulip Data: %s\n", dencode)

	var fullURL = z.BaseURL
	if !strings.HasSuffix(z.BaseURL, "api/v1/messages") {
		fullURL = fmt.Sprintf("%s%s", z.BaseURL, "api/v1/messages")
	}

	req, rerr := http.NewRequest("POST", fullURL, strings.NewReader(dencode))
	if rerr != nil {
		return rerr
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(z.Username, z.Token)

	DebugOut.Printf("Zulip Request: %+v\n", req)

	// TODO: Deadletter handling
	send := func() error {
		var (
			resp *http.Response
		)
		resp, err = http.DefaultClient.Do(req)
		DebugOut.Printf("Zulip Response: %+v\n", resp)
		b, rerr := ioutil.ReadAll(resp.Body)
		DebugOut.Printf("Zulip Response Body: %s\n", string(b))
		resp.Body.Close()

		if err == nil && rerr == nil {
			if resp.StatusCode >= 500 {
				return fmt.Errorf("%d %s", resp.StatusCode, resp.Status)
			}

			// Huzzah
			return nil
		}

		// Fail!
		if err != nil {
			DebugOut.Printf("Zulip Send Failed because '%v'\n", err)
			return err
		}
		DebugOut.Printf("Zulip Send Failed because '%v'\n", rerr)
		return rerr
	}

	r := retrier.New(retrier.ConstantBackoff(z.Retries, z.Interval), nil)

	err = r.Run(send)

	return
}

// writer is an io.Writer
type writer struct {
	Z      *Zulip
	Stream string
	Topic  string
}

// Write sends the []byte to the pre-specified Zulip stream & topic, returning the number of
// bytes probably sent or an error
func (z *writer) Write(p []byte) (n int, err error) {
	err = z.Z.Send(z.Stream, z.Topic, string(p))
	if err == nil {
		n = len(p)
	}
	return
}
