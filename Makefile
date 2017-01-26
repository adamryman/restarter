
docker:
	go build -o fsrestarter ./cmd/fsrestarter
	docker build -t adamryman/fsrestarter .
	rm ./fsrestarter


docker-test:
	go build -o run github.com/adamryman/helloworlddoer
	docker build -t adamryman/fsrestarter-test -f Dockerfile.test .
	rm run
	docker run -d --name fsrestarter-test adamryman/fsrestarter-test
	docker logs fsrestarter-test
	sleep 5
	docker stop fsrestarter-test
	docker logs fsrestarter-test
	docker rm fsrestarter-test


.PHONY: docker docker-test
