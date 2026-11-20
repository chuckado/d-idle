# Static receiver image (stdlib only). Build from the repo root:
#   docker build -f cmd/d2supportrecv/Dockerfile -t didle-support-recv .
FROM golang:1.27 AS build
WORKDIR /src
COPY go.mod ./
COPY cmd/d2supportrecv ./cmd/d2supportrecv
RUN CGO_ENABLED=0 go build -trimpath -ldflags=-s -o /d2supportrecv ./cmd/d2supportrecv

FROM scratch
COPY --from=build /d2supportrecv /d2supportrecv
USER 99:100
EXPOSE 8899
ENTRYPOINT ["/d2supportrecv", "-listen", ":8899", "-dir", "/data/inbox", "-tokens", "/data/tokens", "-trust-proxy", "-proxy-secret-file", "/data/proxy-secret"]
