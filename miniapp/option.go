package miniapp

import (
	"log"
	"net/url"
)

// WithTokenServer 设置Token server
func WithTokenServer(uri string) OptionFunc {
	return func(c *WXMiniClient) {
		var err error
		c.tokenServerURL, err = url.Parse(uri)
		if err != nil {
			log.Fatalln(err)
		}
	}
}
