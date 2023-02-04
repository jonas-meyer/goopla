package goopla

import (
	"net/url"

	"github.com/spf13/viper"
)

type Opt func(*Client) error

func WithBaseURL(u string) Opt {
	return func(c *Client) error {
		baseURL, err := url.Parse(u)
		if err != nil {
			return err
		}
		c.BaseURL = baseURL
		return nil
	}
}

func FromEnv(c *Client) error {
	viper.SetEnvPrefix("zoopla")
	err := viper.BindEnv("api_key")
	if err != nil {
		return err
	}
	c.Credentials.ApiKey = viper.GetString("api_key")
	return nil
}
