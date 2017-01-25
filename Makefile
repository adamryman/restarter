
docker:
	go build -o fsrestarter ./cmd/fsrestarter
	docker build -t adamryman/fsrestarter .
	rm ./fsrestarter


docker-test:
	docker run --name fsrestarter-test --rm adamryman/fsrestarter -d /bin -b ping 0.0.0.0
	docker stop fsrestarter-test


.PHONY: docker docker-test
