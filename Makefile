build:
	@CGO_ENABLED=0 GOOS=linux go build -a -o bin/rd .
install:
	@CGO_ENABLED=0 GOOS=linux go install -a .
