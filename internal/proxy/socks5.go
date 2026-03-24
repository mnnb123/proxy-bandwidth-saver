package proxy

import (
	"context"
	"log"
	"net"

	"github.com/armon/go-socks5"
)

// RemoteResolver prevents DNS leaks by deferring resolution to upstream proxy
type RemoteResolver struct{}

func (r *RemoteResolver) Resolve(ctx context.Context, name string) (context.Context, net.IP, error) {
	return ctx, nil, nil
}

// NewSOCKS5Server creates a SOCKS5 server with optional auth and IP whitelist.
func NewSOCKS5Server(auth *ProxyAuth) (*socks5.Server, error) {
	conf := &socks5.Config{
		Resolver: &RemoteResolver{},
		Logger:   log.Default(),
	}

	if auth != nil {
		auth.mu.RLock()
		authEnabled := auth.authEnabled
		username := auth.username
		password := auth.password
		whitelistEnabled := auth.whitelistEnabled
		auth.mu.RUnlock()

		if authEnabled && username != "" {
			creds := socks5.StaticCredentials{
				username: password,
			}
			conf.AuthMethods = []socks5.Authenticator{
				socks5.UserPassAuthenticator{Credentials: creds},
			}
		}

		if whitelistEnabled {
			conf.Rules = &socks5IPRule{auth: auth}
		}
	}

	return socks5.New(conf)
}

// socks5IPRule implements socks5.RuleSet for IP whitelist checking.
type socks5IPRule struct {
	auth *ProxyAuth
}

func (r *socks5IPRule) Allow(ctx context.Context, req *socks5.Request) (context.Context, bool) {
	srcAddr := req.RemoteAddr.String()
	if !r.auth.CheckIP(srcAddr) {
		log.Printf("SOCKS5: blocked connection from %s (not in whitelist)", srcAddr)
		return ctx, false
	}
	return ctx, true
}
