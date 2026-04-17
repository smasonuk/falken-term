package bootstrap

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

type transport struct {
	underlying http.RoundTripper
	logger     *log.Logger
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("x-portkey-provider", "@openai-aifoundry-swc-001")

	if t.logger != nil {
		t.logger.Printf("LLM Request: %s %s", req.Method, req.URL.String())
		if req.Body != nil {
			body, err := io.ReadAll(req.Body)
			if err == nil {
				t.logger.Printf("LLM Request Payload: %s", string(body))
				req.Body = io.NopCloser(bytes.NewBuffer(body))
			}
		}
	}

	resp, err := t.underlying.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	if t.logger != nil {
		t.logger.Printf("LLM Response Status: %s", resp.Status)
		resp.Body = &loggingReadCloser{ReadCloser: resp.Body, logger: t.logger}
	}

	return resp, nil
}

type loggingReadCloser struct {
	io.ReadCloser
	logger *log.Logger
}

func (l *loggingReadCloser) Read(p []byte) (n int, err error) {
	n, err = l.ReadCloser.Read(p)
	if n > 0 && l.logger != nil {
		l.logger.Printf("LLM Response Chunk: %s", string(p[:n]))
	}
	return n, err
}

func NewOpenAIClient(portkeyAPIKey string, logger *log.Logger) *openai.Client {
	config := openai.DefaultConfig(portkeyAPIKey)
	config.BaseURL = "https://portkey.syngenta.com/v1"
	config.HTTPClient = &http.Client{
		Transport: &transport{
			underlying: http.DefaultTransport,
			logger:     logger,
		},
		Timeout: 5 * time.Minute,
	}
	return openai.NewClientWithConfig(config)
}
