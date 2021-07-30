FROM golang:1.16

WORKDIR $GOPATH/src/github.com/edoardottt/cariddi

COPY . .
RUN go install -v ./...

ENTRYPOINT ["cariddi"]
