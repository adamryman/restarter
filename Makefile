
docker:
	go build -o fsrestarter ./cmd/fsrestarter
	docker build -t adamryman/fsrestarter .
	rm ./fsrestarter


docker-test:
	mkdir -p ./test/target
	go build -o ./test/target/run github.com/adamryman/restarter/test/stdouter
	docker-compose up -d
	docker-compose logs stdouter


.PHONY: docker docker-test
