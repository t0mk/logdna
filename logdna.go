package logdna

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// IngestBaseURL is the base URL for the LogDNA ingest API.
const IngestBaseURL = "https://logs.logdna.com/logs/ingest"

// Config is used by NewClient to configure new clients.
type Config struct {
	APIKey   string
	Hostname string
	Env      string
	App      string
}

// Client is a client to the LogDNA logging service.
type Client struct {
	config  Config
	payload payloadJSON
	apiURL  url.URL
	sync.Mutex
	interval   time.Duration
	stopSignal chan struct{}
}

// logLineJSON represents a log line in the LogDNA ingest API JSON payload.
type logLineJSON struct {
	Timestamp int64  `json:"timestamp"`
	Line      string `json:"line"`
	App       string `json:"app"`
	Env       string `json:"env"`
	Lvl       string `json:"level"`
}

// payloadJSON is the complete JSON payload that will be sent to the LogDNA
// ingest API.
type payloadJSON struct {
	Lines []logLineJSON `json:"lines"`
}

// makeIngestURL creats a new URL to the a full LogDNA ingest API endpoint with
// API key and requierd parameters.
func makeIngestURL(cfg Config) url.URL {
	u, _ := url.Parse(IngestBaseURL)

	u.User = url.User(cfg.APIKey)
	values := url.Values{}
	values.Set("hostname", cfg.Hostname)
	values.Set("now", strconv.FormatInt(time.Time{}.UnixNano(), 10))
	u.RawQuery = values.Encode()

	return *u
}

// NewClient returns a Client configured to send logs to the LogDNA ingest API.
func NewClient(cfg Config) *Client {
	var client Client

	client.config = cfg
	client.interval = 5 * time.Second

	return &client
}

func (c *Client) Run() {
	ticker := time.NewTicker(c.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				c.Flush()
			case <-c.stopSignal:
				ticker.Stop()
				return
			}
		}
	}()
}

const (
	DbgL = "Debug"
	TraL = "Trace"
	InfL = "Info"
	WarL = "Warn"
	ErrL = "Error"
	FtlL = "Fatal"
)

func (c *Client) Dbg(msg ...string) {
	s := strings.Join(msg, " ")
	c.Log(time.Now().UTC(), DbgL, s)
}

func (c *Client) Tra(msg ...string) {
	s := strings.Join(msg, " ")
	c.Log(time.Now().UTC(), TraL, s)
}

func (c *Client) Inf(msg ...string) {
	s := strings.Join(msg, " ")
	c.Log(time.Now().UTC(), InfL, s)
}

func (c *Client) War(msg ...string) {
	s := strings.Join(msg, " ")
	c.Log(time.Now().UTC(), WarL, s)
}

func (c *Client) Err(msg ...string) {
	s := strings.Join(msg, " ")
	c.Log(time.Now().UTC(), ErrL, s)
}

func (c *Client) Ftl(msg ...string) {
	s := strings.Join(msg, " ")
	c.Log(time.Now().UTC(), FtlL, s)
}

func (c *Client) Log(t time.Time, lvl, msg string) {
	// Ingest API wants timestamp in milliseconds so we need to round timestamp
	// down from nanoseconds.
	logLine := logLineJSON{
		Timestamp: t.UnixNano() / int64(time.Millisecond),
		Line:      msg,
		App:       c.config.App,
		Env:       c.config.Env,
		Lvl:       lvl,
	}
	c.Lock()
	c.payload.Lines = append(c.payload.Lines, logLine)
	c.Unlock()
}

// Size returns the number of lines waiting to be sent.
func (c *Client) Size() int {
	return len(c.payload.Lines)
}

// Flush sends any buffered logs to LogDNA and clears the buffered logs.
func (c *Client) Flush() error {
	// Return immediately if no logs to send
	si := c.Size()
	if si == 0 {
		return nil
	}
	c.Lock()
	jsonPayload, err := json.Marshal(c.payload)
	if err != nil {
		return err
	}
	c.Unlock()

	jsonReader := bytes.NewReader(jsonPayload)

	apiURL := makeIngestURL(c.config)

	resp, err := http.Post(apiURL.String(), "application/json", jsonReader)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	c.payload = payloadJSON{}

	return nil
}

// Close closes the client. It also sends any buffered logs.
func (c *Client) Close() error {
	return c.Flush()
}
