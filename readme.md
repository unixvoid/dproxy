### redis structures
upstream:test1:ip
upstream:test1:port

### todo
add upstream wildcard / support more regex expressions

### rkt cleanup
to cleanup unused/stored rkt images use:
`sudo rkt image rm $(sudo rkt image list --fields=id --no-legend)`. this will do
a rkt image rm on all image id's
