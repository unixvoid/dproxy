FROM scratch

ENV GIT_HASH=<GIT_HASH>
ADD rootfs.tar.gz /
COPY run.sh /
COPY dproxy /

CMD ["/run.sh"]
