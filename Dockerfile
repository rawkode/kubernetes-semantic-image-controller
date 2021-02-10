FROM golang:alpine

WORKDIR /code
COPY . /code

RUN go build ./...

ENTRYPOINT [ "/code/kubernetes-semantic-version" ]
