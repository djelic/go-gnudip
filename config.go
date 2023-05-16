package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type config struct {
	Addr       string
	ServerKey  string
	Username   string
	Password   string
	Domains    map[string]string
	Aliases    map[string][]string
	CfApiToken string
}

func newConfig() (*config, error) {
	var c config
	c.Addr = os.Getenv("LISTEN_ADDR")
	if c.Addr == "" {
		c.Addr = ":3495"
	}
	c.ServerKey = os.Getenv("SERVER_KEY")
	if c.ServerKey == "" {
		return nil, errors.New("SERVER_KEY not set")
	}
	c.Username = os.Getenv("SERVER_USERNAME")
	if c.Username == "" {
		return nil, errors.New("SERVER_USERNAME not set")
	}
	c.Password = os.Getenv("SERVER_PASSWORD")
	if c.Password == "" {
		return nil, errors.New("SERVER_PASSWORD not set")
	}
	c.Domains = make(map[string]string)
	domainHandlers := strings.Split(os.Getenv("DOMAINS"), ";")
	for _, domainHandler := range domainHandlers {
		dh := strings.Split(domainHandler, ":")
		if len(dh) != 2 {
			errMsg := fmt.Sprintf("DOMAINS should be in format `domain:handler`, `%s` provided", domainHandler)
			return nil, errors.New(errMsg)
		}
		c.Domains[dh[0]] = dh[1]
	}
	aliases := os.Getenv("ALIASES")
	if aliases != "" {
		c.Aliases = make(map[string][]string)
		domainAliases := strings.Split(aliases, ";")
		for _, domainAlias := range domainAliases {
			da := strings.Split(domainAlias, ":")
			if len(da) != 2 {
				errMsg := fmt.Sprintf("ALIASES should be in format `domain:alias1,alias2`, `%s` provided", domainAlias)
				return nil, errors.New(errMsg)
			}
			aliases := strings.Split(da[1], ",")
			c.Aliases[da[0]] = aliases
		}
	}
	c.CfApiToken = os.Getenv("CLOUDFLARE_API_TOKEN")
	return &c, nil
}
