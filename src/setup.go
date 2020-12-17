package syntropy

import (
	"errors"
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"time"
)

func init() { plugin.Register("syntropy", setup) }

func setup(c *caddy.Controller) error {
	syn, err := newSyntropy(c)

	if err != nil {
		log.Fatalf("Failed to initialize Syntropy %v", err)
	}

	token := login(syn.Url, syn.Username, syn.Password)
	syn.AccessToken = token

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		syn.Next = next
		return syn
	})

	return nil
}

// credits to netbox code for parsing arguments
func newSyntropy(c *caddy.Controller) (Syntropy, error) {
	url := ""
	username := ""
	password := ""
	localCacheDuration := ""
	ttl := time.Second
	var err error

	for c.Next() {
		if c.NextBlock() {
			for {
				switch c.Val() {
				case "url":
					if !c.NextArg() {
						c.ArgErr()
					}
					url = c.Val()
				case "username":
					if !c.NextArg() {
						c.ArgErr()
					}
					username = c.Val()
				case "password":
					if !c.NextArg() {
						c.ArgErr()
					}
					password = c.Val()
				case "localCacheDuration":
					if !c.NextArg() {
						c.ArgErr()
					}
					localCacheDuration = c.Val()
					ttl, err = time.ParseDuration(localCacheDuration)
					if err != nil {
						localCacheDuration = ""
					}
				}

				if !c.Next() {
					break
				}
			}
		}
	}

	if url == "" || username == "" || password == "" || localCacheDuration == "" {
		return Syntropy{}, errors.New("Failed to parse Syntropy DNS config")
	}

	return Syntropy{
		Url:      url,
		Username: username,
		Password: password,
		Ttl:      ttl,
	}, nil
}
