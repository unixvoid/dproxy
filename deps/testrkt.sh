#!/bin/bash
CURRENT_PATH=$(pwd)

sudo rkt run \
	--insecure-options=all \
        --volume redis,kind=host,source=/tmp/ \
        --volume config,kind=host,source=$CURRENT_PATH/config.gcfg \
        --volume upstream,kind=host,source=$CURRENT_PATH/upstream/ \
	--port=dns-tcp:53 \
	--port=dns-udp:53 \
	--debug \
        unixvoid.com/dproxy

#CURRENT_DIR=$(pwd)
#--port=dns-tcp:8053 \
#--port=dns-udp:8053 \
#--volume redis,kind=host,source=$CURRENT_DIR \
#--net=host \
