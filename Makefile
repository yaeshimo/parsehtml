COMMIT := $(shell git rev-parse --verify --short HEAD)
DATE := $(shell date -u +%Y-%m-%d)
LDFLAGS := -ldflags " \
	-X 'main.Commit=$(COMMIT)' \
	-X 'main.Date=$(DATE)' \
	"

all : build

build :
	go build $(LDFLAGS)

test :
	go test -race -v

run : build
	./parsehtml -html testdata/ignore/test.html -config testdata/ignore/config.json

arm7 :
	GOOS=linux GOARCH=arm GOARM=7 go build

win :
	GOOS=windows GOARCH=amd64 go build
