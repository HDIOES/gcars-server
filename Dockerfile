FROM golang
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
WORKDIR $GOPATH/src/github.com/HDIOES/gcars-server
COPY Gopkg.toml Gopkg.lock ./
COPY . ./
RUN dep ensure
RUN go install github.com/HDIOES/gcars-server
RUN cp configuration.json $GOPATH/bin/
RUN cp -r migrations/ $GOPATH/bin/
WORKDIR $GOPATH/bin
ENTRYPOINT ["./gcars-server", "dbmode"]