FROM golang:latest 
 
WORKDIR /usr/local/go/src/main

COPY ./ /usr/local/go/src/main

RUN go mod download && go build -o main .

EXPOSE 8181

ENTRYPOINT ["/usr/local/go/src/main/main"]