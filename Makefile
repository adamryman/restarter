
docker:
	go build -o fsrestarter ./cmd/fsrestarter
	docker build -t adamryman/fsrestarter .
	rm ./fsrestarter

.PHONY: docker
