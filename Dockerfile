FROM golang:alpine as builder
RUN mkdir /build && apk add --update git
ADD . /build/
WORKDIR /build 
RUN set -x && \
  cd /build && \
  go get -t -v k8s.io/apimachinery/pkg/apis/meta/v1 && \
  go get -t -v k8s.io/client-go/kubernetes && \
  go get -t -v k8s.io/client-go/rest && \
  go get -t -v k8s.io/client-go/tools/clientcmd && \
  CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o kube-dependency-controller .

FROM scratch
COPY --from=builder /build/kube-dependency-controller /app/
WORKDIR /app
CMD ["./kube-dependency-controller"]
