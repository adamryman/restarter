
docker:
	go build ./cmd/fsrestarter
	docker build -t adamryman/fsrestarter .
	rm fsrestarter

.PHONY: docker
