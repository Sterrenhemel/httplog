# syntax=docker/dockerfile:1
FROM golang:1.18-buster as builder

ARG TARGET=http

RUN --mount=type=secret,id=GO_PRIVATE_TOKEN \
  GO_PRIVATE_TOKEN=$(cat /run/secrets/GO_PRIVATE_TOKEN) && \
  git config --global url."https://Sterrenhemel:${GO_PRIVATE_TOKEN}@github.com".insteadOf "https://github.com"

ENV GO111MODULE=on \
    GOPRIVATE=github.com/Sterrenhemel/*

WORKDIR /usr/src/app

ADD src/go.mod .
ADD src/go.sum .
RUN go mod download

#
COPY . .
RUN build/build.sh ${TARGET}

# deploy run
FROM alpine

COPY --from=Builder /usr/src/app/deploy /app

ENV ENV="prod"
ENV SERVICE_NAME=${TARGET}

WORKDIR /app

CMD ["sh", "./bootstrap.sh"]