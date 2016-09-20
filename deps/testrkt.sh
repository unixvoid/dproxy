#!/bin/bash
CURRENT_PATH=$(pwd)

sudo rkt run \
	--insecure-options=all \
	--net=host \
        --volume redis,kind=host,source=/tmp/ \
        --volume config,kind=host,source=$CURRENT_PATH/config.gcfg \
        --volume upstream,kind=host,source=$CURRENT_PATH/upstream/ \
	--debug \
        ./dproxy.aci

#CURRENT_DIR=$(pwd)
#--port=dns:8053 \
#--volume redis,kind=host,source=$CURRENT_DIR \
#--net=host \
