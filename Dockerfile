#
# Build
#

FROM golang:1.12-alpine AS builder

ENV GO_DOMAIN="github.com" \
    GO_GROUP="otaviof" \
    GO_PROJECT="vault-handler"

ENV APP_DIR="${GOPATH}/src/${GO_DOMAIN}/${GO_GROUP}/${GO_PROJECT}"

RUN apk --update add git make
RUN go get -u github.com/golang/dep/cmd/dep

RUN mkdir -v -p ${APP_DIR}
WORKDIR ${APP_DIR}
COPY Makefile Gopkg.* ./
RUN make bootstrap
COPY . ./
RUN make

#
# Run
#

FROM golang:1.12-alpine

ENV GO_DOMAIN="github.com" \
    GO_GROUP="otaviof" \
    GO_PROJECT="vault-handler"

ENV APP_DIR="${GOPATH}/src/${GO_DOMAIN}/${GO_GROUP}/${GO_PROJECT}"

RUN apk --update add bash
COPY --from=builder ${APP_DIR}/build/${GO_PROJECT} /usr/local/bin/${GO_PROJECT}

WORKDIR /
ENTRYPOINT ["/usr/local/bin/${GO_PROJECT}"]
