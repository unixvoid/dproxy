package main

import (
	"net"

	"github.com/miekg/dns"
	"github.com/unixvoid/glogger"
	"gopkg.in/redis.v3"
)

func resolve(w dns.ResponseWriter, req *dns.Msg, redisClient *redis.Client) {
	hostname := req.Question[0].Name
	//glogger.Cluster.Println(hostname)
	//domain := parseHostname(hostname)

	// ditch the request if its infinitely recursive
	if hostname == "." {
		return
	}
	// check the domain to see if it is in redis
	err, upstream := checkDomain(redisClient, hostname)
	if err != nil {
		//glogger.Debug.Println("response from redis: ", err)
		if config.Dproxy.UseMasterUpstream {
			upstream = config.Dproxy.MasterUpstream
		} else {
			// if we are set not to use upstream, return a RXCODE8
			glogger.Debug.Println("ipv4 entry not found in records, sending rcode3")

			rr := new(dns.A)
			rr.Hdr = dns.RR_Header{Name: hostname, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 1}
			rr.A = net.ParseIP("")

			// craft reply
			rep := new(dns.Msg)
			rep.SetReply(req)
			rep.SetRcode(req, dns.RcodeNameError)
			rep.Answer = append(rep.Answer, rr)

			// send it
			w.WriteMsg(rep)
			return
		}
	}
	glogger.Debug.Printf("routing request %s to %s\n", hostname, upstream)

	transport := "udp"
	if _, ok := w.RemoteAddr().(*net.TCPAddr); ok {
		transport = "tcp"
	}
	c := &dns.Client{Net: transport}
	resp, _, err := c.Exchange(req, upstream)
	if err != nil {
		glogger.Debug.Println(err)
		dns.HandleFailed(w, req)
		return
	}

	w.WriteMsg(resp)
	return
}
