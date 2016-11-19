install: dep
	go install

test:
	go test -v

test_with_docker:
	docker run -v $(PWD):/go/src/github.com/pocke/ptmux ptmux

docker_build:
	docker build -t ptmux:latest .

dep:
	glide install
