package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/unixvoid/glogger"
	"gopkg.in/gcfg.v1"
)

type Config struct {
	Cryo struct {
		Loglevel          string
		Port              int
		UpstreamLocation  string
		UpstreamExtension string
	}
}

var (
	config = Config{}
)

func main() {
	readConf()
	initLogger(config.Cryo.Loglevel)
	listUpstreams()
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

func listUpstreams() {
	//dirname := "." + string(filepath.Separator)
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

							// THROW THIS SHIT IN REDIS
							fmt.Printf("upstream:%s:%s :: %s\n", entryName, field, value)
						}
					}
				}
			}
		}
	}
}

func parseString(line string) (error, string, string) {
	var s []string
	var tmpStr string
	var field string
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
