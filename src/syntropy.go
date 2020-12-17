package syntropy

import (
	"context"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	"net"
	"strings"
	"time"
)

var log = clog.NewWithPlugin("syntropy")

type Syntropy struct {
	AccessToken string
	Url         string
	Username    string
	Password    string
	Ttl         time.Duration
	Next        plugin.Handler
}

func (s Syntropy) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	answers := []dns.RR{}
	state := request.Request{W: w, Req: r}

	name := strings.TrimRight(state.QName(), ".")
	ips := query(name, s.Url, s.AccessToken, s.Ttl)

	if len(ips) == 0 {
		return plugin.NextOrFailure(s.Name(), s.Next, ctx, w, r)
	}

	for _, ip := range ips {
		rec := new(dns.A)
		rec.Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 3600}
		rec.A = net.ParseIP(ip)
		answers = append(answers, rec)
	}

	m := new(dns.Msg)
	m.Answer = answers
	m.SetReply(r)
	w.WriteMsg(m)

	return dns.RcodeSuccess, nil
}

func (s Syntropy) Name() string { return "syntropy" }
