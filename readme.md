# DProxy
[![Build Status (Travis)](https://travis-ci.org/unixvoid/dproxy.svg?branch=master)](https://travis-ci.org/unixvoid/dproxy)  
dproxy is a dns proxy written in golang. The proxy runs at a dns level and is
able to route specific DNS traffic to multiple backend servers.  
This tool was designed to run multiple different dns servers on one box.  For
example if you wanted to run 3 nameservers completely independent from
each other, toss dproxy up and point the configuration to the backend
nameservers.  


## Runnign dproxy
There are 3 main ways to run dproxy:

1. **From Source**: To run dproxy from source we need to pull the required
   golang depdencies and then run.  We can accomplish this with make.  
   `make dependencies` and then `make run`.  If you want to statically compile
   dproxy for portability use `make stat` instead of `make run` and the
   statically compiled binary will be produced in `bin/`.

2. **ACI/rkt**: We have public rkt images hosted on unixvoid.com. You can
   download the image from [here](https://cryo.unixvoid.com/bin/rkt/cryodns) or
   give us a fetch with `rkt fetch unixvoid.com/cryodns`.  This image can be run
   with rkt or you can grab our [service file](https://github.com/unixvoid/dproxy/blob/master/deps/dproxy.service) and run with systemd.

3. **Docker**: We have dproxy pre-packaged over on the
   [dockerhub](https://hub.docker.com/r/unixvoid/dproxy), go grab the latest and
   run: `docker run -d -p 53:53 unixvoid/dproxy`.  


## Server Configuration
The server configuration is pretty straightforward, the following is the default
configuration that dproxy ships with.
```
[dproxy]						# this section is the start of the server main config.
	loglevel		= "debug"		# loglevel, this can be one of [debug, cluster, info, error]
	port			= 53			# port for the DNS server to listen on
	upstreamlocation	= "upstream/"		# the location for all upstream server configs. (see upstream configuration)
	upstreamextension	= ".prox"		# file extention that dproxy will parse for upstreams
	usemasterupstream	= true			# unsed right now, we will add in a way to return a non-authoritative DNS response when set to flase
	masterupstream		= "8.8.8.8:53"		# upstream DNS server to use when entry is not found locally

[redis]							# this section is the start of the redis configuration
	host			= "localhost:6379"	# ip and port to connect to redis on
	password		= ""			# password for the redis server (if it has one)
```

## Upstream Configuration
dproxy uses an INI style configuration for configuring upstream proxies.  dproxy
parses these configurations at runtime and effects will not be implimented until
the application is restarted.  These upstream files can be all crammed into one
file, split into different files per domain, or a mix of both.  By default all
domain configs are stored in the directory `upstream/`, and all of the files
must have the file extention `.prox`.  Both of these settings (directory and
extension) can be set in the main server configuration file.  The following is 
part what I use on my production stack.  

`upstream/unixvoid.prox`  
```
[*.unixvoid.com]
	address		= 192.168.1.9
	port		= 8054

[unixvoid.com]
	address		= 192.168.1.9
	port		= 8054
```  
  
`upstream/cryo.network.prox`  
```
[*.cryo.network]
	address		= 192.168.1.9
	port		= 8053

[cryo.network]
	address		= 192.168.1.9
	port		= 8053
```  
In these configs I use two different DNS servers running on the backend.  One
for `unixvoid.com` and one for `cryo.network`.  In each of these nameservers I
redirect any subdomain ie `*.unixvoid.com` and the domain itself ie
`unixvoid.com`.  You can also set specific subdomains if you want to.  There is
no limit to how much is parsed or putting more than one domain in one file.  For
instance I could have both of these domains `unixvoid.com` and `cryo.network`
into one config, but instead I chose to break them into 2 different files for
portability.


### rkt cleanup
to cleanup unused/stored rkt images use:
`sudo rkt image rm $(sudo rkt image list --fields=id --no-legend)`. this will do
a rkt image rm on all image id's
