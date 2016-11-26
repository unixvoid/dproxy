package main

import (
	"fmt"
	"io/ioutil"
	"os"

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
		UseMasterUpstream bool
		MasterUpstream    string
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
