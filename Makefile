export VERSION = ${shell cat VERSION}

all:
	go build -ldflags "-X github.com/osm/qwbs/internal/version.version=${VERSION}'"

clean:
	rm -f qwbs
