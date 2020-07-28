FROM golang:rc-alpine3.12

RUN apk add git
RUN cd / && git clone https://github.com/scottbrumley/hdctl.git
#ADD config.json /hdctl/config.json
WORKDIR /hdctl
ENV GOPATH=/hdctl/
RUN go get "github.com/eclipse/paho.mqtt.golang"
RUN go get "github.com/gorilla/mux"
RUN go get "go.mongodb.org/mongo-driver/bson"
RUN go get "go.mongodb.org/mongo-driver/mongo"
RUN go get "go.mongodb.org/mongo-driver/mongo/options"
RUN go build *.go

CMD /hdctl/hdctl