FROM golang:1.17.5-bullseye as builder

WORKDIR /tmp/setup

RUN apt-get update && \
        apt-get install -y libwebkit2gtk-4.0-dev && \
        apt-get autoremove

WORKDIR /app
COPY . .

CMD ["go", "build", "-trimpath", "-ldflags", "-s -w", "-o", "bin/mad", "."]


