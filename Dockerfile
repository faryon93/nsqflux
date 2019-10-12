# ----------------------------------------------------------------------------------------
# Image: Builder
# ----------------------------------------------------------------------------------------
FROM golang:alpine as builder

# setup the environment
ENV TZ=Europe/Berlin

# install dependencies
RUN apk --update --no-cache add git gcc musl-dev tzdata
WORKDIR /work
ADD ./ ./

# build the go binary
RUN go build -ldflags \
        '-X "main.BuildTime='$(date -Iminutes)'" \
         -X "main.GitCommit='$(git rev-parse --short HEAD)'" \
         -X "main.GitBranch='$(git rev-parse --abbrev-ref HEAD)'" \
         -s -w' \
         -v -o /tmp/nsqflux .
RUN chown nobody:nobody /tmp/nsqflux && \
    chmod +x /tmp/nsqflux

# ----------------------------------------------------------------------------------------
# Image: Deployment
# ----------------------------------------------------------------------------------------
FROM alpine:latest
MAINTAINER Maximilian Pachl <m@ximilian.info>

RUN apk --update --no-cache add ca-certificates

# add relevant files to container
COPY --from=builder /tmp/nsqflux /usr/sbin/nsqflux

USER nobody
CMD /usr/sbin/nsqflux
