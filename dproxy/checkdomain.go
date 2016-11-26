package main

import (
	"fmt"
	"strings"

	"github.com/unixvoid/glogger"
	"gopkg.in/redis.v3"
)

func checkDomain(redisClient *redis.Client, domainName string) (error, string) {
	var (
		err     error
		address string
		port    string
	)
	// fully qualify the domain name if it is not already:
	if string(domainName[len(domainName)-1]) != "." {
		domainName = fmt.Sprintf("%s.", domainName)
	}

	// first check the number of '.'s
	// if there are more than 2 it is a subdomain. Check for root domain's wildcard
	// before we search for the domain. ie mail.google.com. : look for *.google.com first
	subd := strings.Count(domainName, ".")
	if subd > 2 {
		tmpSplit := strings.Split(domainName, ".")

		wildcardDomain := fmt.Sprintf("*.%s.%s.", tmpSplit[(len(tmpSplit)-3)], tmpSplit[(len(tmpSplit)-2)])
		address, err = redisClient.Get(fmt.Sprintf("upstream:%s:address", wildcardDomain)).Result()
		if err == nil {
			port, _ = redisClient.Get(fmt.Sprintf("upstream:%s:port", wildcardDomain)).Result()
			//glogger.Debug.Printf("returning wildcard '%s:%s' to client\n", address, port)
			return nil, fmt.Sprintf("%s:%s", address, port)
		}
	}

	glogger.Debug.Printf("checking redis for: upstream:%s:address\n", domainName)
	address, err = redisClient.Get(fmt.Sprintf("upstream:%s:address", domainName)).Result()
	port, err = redisClient.Get(fmt.Sprintf("upstream:%s:port", domainName)).Result()
	if err != nil {
		glogger.Debug.Println("data not found in db")
		return fmt.Errorf("data not found in db"), ""
	}
	return nil, fmt.Sprintf("%s:%s", address, port)
}
