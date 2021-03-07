# Kubernetes Semantic Image Controller

This is a Kubernetes Mutating Webhook Controller developed as part of the [Rawkode Live Episode on 
Writing a Kubernetes Controller](https://www.youtube.com/watch?v=RLpzsAQtZ7M)

This is mostly prototyping code developed and tested against Kubernetes 1.20. It may not be quite ready for
production usage. Use at your own risk!

## What this controller does

In your manifest, you typically specify a full image path, like so
```yaml
    image: nginx:1.19.7
```

What if you wanted to be more liberal with the versioning, wouldn't be nice to pick up bug fixes and improvements
as pods rotate?

What if you could provide a version range like so?

```yaml
    image: "nginx: >= 1.19, <= 1.20"
```

This webhook takes the semantic version constraint and resolves it into the latest version that satisfies that
constraint as part of a [Kubernetes Mutating Webhook](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/)

## Building

A Makefile and Dockerfile has been provided to build the Controller into a Docker Image. This will compile a
Linux binary and put it into a Docker Image tagged `kubernetes-semantic-image-controller:latest`

```sh
$ make clean
$ make build
```

## Testing

You can run the tests using `go test`

```sh
$ go test -v ./...
```
