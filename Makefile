build:
	@CGO_ENABLED=0 GOOS=linux go build -a -o bin/fdup .
install:
	@CGO_ENABLED=0 GOOS=linux go install -a .
