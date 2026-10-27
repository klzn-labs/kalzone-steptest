FROM golang:1.25-alpine AS build
ARG VERSION=dev
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /steptest .

FROM alpine:3.22
COPY --from=build /steptest /steptest
EXPOSE 8080
ENTRYPOINT ["/steptest"]
