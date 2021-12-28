FROM golang:1.17.4-alpine3.15 as build
RUN apk add --no-cache gcc libc-dev ca-certificates && update-ca-certificates
WORKDIR /app

ENV CGO_ENABLED=0
ENV GO111MODULE=on
#ENV GOFLAGS=-mod=vendor

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o /app/flook-go .

FROM scratch AS final
LABEL maintainer="D3v <mark@zsibok.hu>"
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /app/flook-go /


CMD [ "./flook-go" ]
