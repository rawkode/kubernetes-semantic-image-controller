build:
	GOOS=linux go build -o kubernetes-semantic-image-controller-linux .
	docker build -t kubernetes-semantic-image-controller:latest .
	rm -rf kubernetes-semantic-image-controller-linux

clean:
	rm -rf kubernetes-semantic-image-controller-linux
