FROM golang:buster
RUN go get -u github.com/golang/dep/cmd/dep 
ADD Gopkg.toml /go/src/github.com/valentin2105/k8s-mattermost/
ADD Gopkg.lock /go/src/github.com/valentin2105/k8s-mattermost/
WORKDIR /go/src/github.com/valentin2105/k8s-mattermost
RUN dep ensure -vendor-only
ADD *.go /go/src/github.com/valentin2105/k8s-mattermost/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o k8s-mattermost -v

FROM alpine:latest
ENV KUBE_LATEST_VERSION="v1.16.1"
RUN  apk update \
     && apk --no-cache add ca-certificates bash curl \
     && curl -L https://storage.googleapis.com/kubernetes-release/release/${KUBE_LATEST_VERSION}/bin/linux/amd64/kubectl -o /usr/local/bin/kubectl \
     && chmod +x /usr/local/bin/kubectl

WORKDIR /

COPY --from=0 /go/src/github.com/valentin2105/k8s-mattermost/k8s-mattermost .
ADD entrypoint.sh .
ADD config.toml.dist .

CMD ["/entrypoint.sh"]
