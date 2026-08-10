FROM alpine:3.20
ARG MESSAGE=default
RUN test "$MESSAGE" = expected
CMD ["cat", "/etc/alpine-release"]
