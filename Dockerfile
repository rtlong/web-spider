FROM google/golang

ADD . /gopath/src/github.com/rtlong/web-spider
WORKDIR /gopath/src/github.com/rtlong/web-spider
RUN go get
RUN go install

ENTRYPOINT ["web-spider"]
