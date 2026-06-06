FROM golang:1.25.6

WORKDIR /researcher

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux \
    go build -o /bin/researcher ./

EXPOSE 8080

ENTRYPOINT ["/bin/researcher"]
