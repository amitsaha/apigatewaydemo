FROM golang:alpine
ADD . /go/src/apigateway
WORKDIR /go/src/apigateway
EXPOSE 8000
RUN go build -o apigateway main.go && go install -v
RUN ls -lrt /go/bin/
CMD /go/bin/apigateway
