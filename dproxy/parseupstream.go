package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/unixvoid/glogger"
	"gopkg.in/redis.v3"
)

func parseUpstreams(redisClient *redis.Client) {
	dirname := fmt.Sprintf("%s", config.Dproxy.UpstreamLocation)

	d, err := os.Open(dirname)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer d.Close()

	files, err := d.Readdir(-1)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// as bad as this looks, its only O(n)
	// open file, parse line by line
	for _, file := range files {
		if file.Mode().IsRegular() {
			if filepath.Ext(file.Name()) == config.Dproxy.UpstreamExtension {
				tmpfile := fmt.Sprintf("%s%s", config.Dproxy.UpstreamLocation, file.Name())
				f, _ := ioutil.ReadFile(tmpfile)
				var entryName string

				lines := strings.Split(string(f), "\n")
				for i := range lines {
					err, field, value := parseString(lines[i])
					if err == nil {
						if field == "server" {
							// fully qualify the domain name if it is not already:
							if string(value[len(value)-1]) != "." {
								value = fmt.Sprintf("%s.", value)
							}
							entryName = value
						} else {
							// add entries to redis
							redisEntry := fmt.Sprintf("upstream:%s:%s", entryName, field)
							// make sure 'redisEntry' is not space padded
							redisEntry = strings.Replace(redisEntry, " ", "", -1)
							glogger.Debug.Printf("setting '%s' to '%s' in redis", redisEntry, value)
							redisClient.Set(redisEntry, value, 0).Err()
						}
					}
				}
			}
		}
	}
}

func parseString(line string) (error, string, string) {
	var (
		s      []string
		tmpStr string
		field  string
	)
	chr := "[ ]"

	if line == "" || strings.Contains(line, "#") {
		return fmt.Errorf("cannot use empty string"), "", ""
	}
	line = strings.Replace(line, "\t", "", -1)

	// check if = exists
	if strings.Contains(line, "=") {
		s = strings.Split(line, "=")
		tmpStr = s[1]
		field = s[0]
	} else {
		tmpStr = line
		field = "server"
	}

	value := strings.Map(func(r rune) rune {
		if strings.IndexRune(chr, r) < 0 {
			return r
		}
		return -1
	}, tmpStr)
	return nil, field, value
}
