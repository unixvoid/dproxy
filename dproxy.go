package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/miekg/dns"
	"github.com/unixvoid/glogger"
	"gopkg.in/gcfg.v1"
	"gopkg.in/redis.v3"
)

type Config struct {
	Cryo struct {
		Loglevel          string
		Port              int
		UpstreamLocation  string
		UpstreamExtension string
	}
	Redis struct {
		Host     string
		Password string
	}
}

var (
	config = Config{}
)

func main() {
	readConf()
	initLogger(config.Cryo.Loglevel)

	redisClient, redisErr := initRedisConnection()
	if redisErr != nil {
		glogger.Error.Println("redis connection cannot be made.")
		glogger.Error.Println("dproxy will continue to function in passthrough mode only")
	} else {
		glogger.Debug.Println("connection to redis succeeded.")
	}

	// read in confs
	parseUpstreams(redisClient)

	// format the string to be :port
	port := fmt.Sprint(":", config.Cryo.Port)

	udpServer := &dns.Server{Addr: port, Net: "udp"}
	tcpServer := &dns.Server{Addr: port, Net: "tcp"}
	glogger.Info.Println("started server on", config.Cryo.Port)
	dns.HandleFunc(".", func(w dns.ResponseWriter, req *dns.Msg) {
		go resolve(w, req, redisClient)
	})

	go func() {
		glogger.Error.Println(udpServer.ListenAndServe())
	}()
	glogger.Error.Println(tcpServer.ListenAndServe())

}

func readConf() {
	// init config file
	err := gcfg.ReadFileInto(&config, "config.gcfg")
	if err != nil {
		panic(fmt.Sprintf("Could not load config.gcfg, error: %s\n", err))
	}
}

func initLogger(logLevel string) {
	// init logger
	if logLevel == "debug" {
		glogger.LogInit(os.Stdout, os.Stdout, os.Stdout, os.Stderr)
	} else if logLevel == "cluster" {
		glogger.LogInit(os.Stdout, os.Stdout, ioutil.Discard, os.Stderr)
	} else if logLevel == "info" {
		glogger.LogInit(os.Stdout, ioutil.Discard, ioutil.Discard, os.Stderr)
	} else {
		glogger.LogInit(ioutil.Discard, ioutil.Discard, ioutil.Discard, os.Stderr)
	}
}
func initRedisConnection() (*redis.Client, error) {
	// init redis connection
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.Redis.Host,
		Password: config.Redis.Password,
		DB:       0,
	})

	_, redisErr := redisClient.Ping().Result()
	return redisClient, redisErr
}

func resolve(w dns.ResponseWriter, req *dns.Msg, redisClient *redis.Client) {
	hostname := req.Question[0].Name
	glogger.Cluster.Println(hostname)
	//domain := parseHostname(hostname)
	transport := "udp"
	if _, ok := w.RemoteAddr().(*net.TCPAddr); ok {
		transport = "tcp"
	}
	c := &dns.Client{Net: transport}
	resp, _, err := c.Exchange(req, "192.168.1.9:5553")
	if err != nil {
		glogger.Debug.Println(err)
		dns.HandleFailed(w, req)
		return
	}

	w.WriteMsg(resp)
	return
}

func parseUpstreams(redisClient *redis.Client) {
	dirname := fmt.Sprintf("%s", config.Cryo.UpstreamLocation)

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
			if filepath.Ext(file.Name()) == config.Cryo.UpstreamExtension {
				tmpfile := fmt.Sprintf("%s%s", config.Cryo.UpstreamLocation, file.Name())
				f, _ := ioutil.ReadFile(tmpfile)
				var entryName string

				lines := strings.Split(string(f), "\n")
				for i := range lines {
					err, field, value := parseString(lines[i])
					if err == nil {
						if field == "server" {
							entryName = value
						} else {

							// add entries to redis
							redisEntry := fmt.Sprintf("upstream:%s:%s", entryName, field)
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
