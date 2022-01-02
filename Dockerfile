FROM golang:1.17.5-bullseye as builder
ARG GOOS=linux
ARG GOARCH=amd64

ENV GOOS=$GOOS
ENV GOARCH=$GOARCH

WORKDIR /tmp/setup

RUN apt-get update && \
        apt-get install -y libwebkit2gtk-4.0-dev && \
        apt-get autoremove

WORKDIR /app
COPY . .

CMD ["go", "build", "-trimpath", "-ldflags", "-s -w", "-o", "bin/mad", "."]


