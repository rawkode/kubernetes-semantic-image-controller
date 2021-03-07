FROM alpine:3.13.1

COPY kubernetes-semantic-image-controller-linux /kubernetes-semantic-image-controller-linux

CMD ["/kubernetes-semantic-image-controller-linux"]
