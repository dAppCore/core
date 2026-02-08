FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod main.go ./
RUN go build -trimpath -ldflags="-w -s" -o /vanity-import .

FROM alpine:3.21
COPY --from=build /vanity-import /vanity-import
EXPOSE 8080
ENTRYPOINT ["/vanity-import"]
