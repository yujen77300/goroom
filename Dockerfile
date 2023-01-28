FROM golang:1.19-alpine3.16
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download
COPY . .


RUN CGO_ENABLED=0 go build -o /bin/app ./cmd

FROM alpine
WORKDIR /src

COPY --from=0 /bin/app /bin/app
COPY --from=0 /src/views /src/views
COPY --from=0 /src/static /src/static
COPY config.env /src/config.env

ENTRYPOINT ["/bin/app"]
