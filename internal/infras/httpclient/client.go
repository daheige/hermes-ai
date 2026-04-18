package httpclient

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

var (
	HTTPClient                   *http.Client
	ImpatientHTTPClient          *http.Client
	UserContentRequestHTTPClient *http.Client
)

// ClientConfig client 配置
type ClientConfig struct {
	UserContentRequestProxy   string
	UserContentRequestTimeout int
	RelayProxy                string
	RelayTimeout              int
}

// Init 初始化http client
func Init(config ClientConfig) {
	if config.UserContentRequestProxy != "" {
		log.Printf(fmt.Sprintf("using %s as proxy to fetch user content\n", config.UserContentRequestProxy))
		proxyURL, err := url.Parse(config.UserContentRequestProxy)
		if err != nil {
			log.Fatalf(fmt.Sprintf("USER_CONTENT_REQUEST_PROXY set but invalid: %s", config.UserContentRequestProxy))
		}
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
		UserContentRequestHTTPClient = &http.Client{
			Transport: transport,
			Timeout:   time.Second * time.Duration(config.UserContentRequestTimeout),
		}
	} else {
		UserContentRequestHTTPClient = &http.Client{}
	}
	var transport http.RoundTripper
	if config.RelayProxy != "" {
		log.Printf(fmt.Sprintf("using %s as api relay proxy\n", config.RelayProxy))
		proxyURL, err := url.Parse(config.RelayProxy)
		if err != nil {
			log.Fatalf(fmt.Sprintf("USER_CONTENT_REQUEST_PROXY set but invalid: %s", config.UserContentRequestProxy))
		}
		transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}

	if config.RelayTimeout == 0 {
		HTTPClient = &http.Client{
			Transport: transport,
		}
	} else {
		HTTPClient = &http.Client{
			Timeout:   time.Duration(config.RelayTimeout) * time.Second,
			Transport: transport,
		}
	}

	ImpatientHTTPClient = &http.Client{
		Timeout:   5 * time.Second,
		Transport: transport,
	}
}
