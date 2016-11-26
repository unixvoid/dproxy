### redis structures
[![Build Status (Travis)](https://travis-ci.org/unixvoid/dproxy.svg?branch=master)](https://travis-ci.org/unixvoid/dproxy)  

### rkt cleanup
to cleanup unused/stored rkt images use:
`sudo rkt image rm $(sudo rkt image list --fields=id --no-legend)`. this will do
a rkt image rm on all image id's
