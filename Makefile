COMMIT := $(shell git rev-parse --verify --short HEAD)
DATE := $(shell date -u +%Y-%m-%d)
LDFLAGS := -ldflags " \
	-X 'main.Commit=$(COMMIT)' \
	-X 'main.Date=$(DATE)' \
	"

all : build

build :
	go build $(LDFLAGS) -- ./cmd/parsehtml

test :
	go test -race -v ./filter

run : build
	./parsehtml -html testdata/test.html -config testdata/config.json

arm7 :
	GOOS=linux GOARCH=arm GOARM=7 go build

win :
	GOOS=windows GOARCH=amd64 go build
