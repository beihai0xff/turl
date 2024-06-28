ARG GO_VERSION
FROM golang:${GO_VERSION} AS builder

ENV GOPROXY=https://goproxy.io,direct
ENV GO111MODULE=on

WORKDIR /go/src/turl

COPY go.mod .
COPY go.sum .
COPY api .

RUN go mod download -x

COPY . /go/src/turl
ARG BUILD_DATE
ARG BUILD_COMMIT
ARG BUILD_VERSION
RUN make build/binary


FROM ubuntu:22.04 AS prod

ARG BUILD_DATE

# first, a bit about this container
LABEL org.smartx.image.created="${BUILD_DATE}" \
      org.smartx.image.authors="beihai <beihai0xff@gmail.com>"

ENV GIN_MODE=release

EXPOSE 8080/tcp

# let docker know how to test container health
HEALTHCHECK CMD /app/turl server health || exit 1

COPY --from=builder /go/src/turl/build/dist/binary/turl  /app/turl

WORKDIR /app
ENTRYPOINT ["/app/turl", "start"]
CMD ["-f", "./config.yaml"]