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
			glogger.Debug.Println(err)
			dns.HandleFailed(w, req)
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
