SHELL := /bin/bash

icarus:
	go run cmd/icarus/main.go

icontent:
	go run cmd/icontent/main.go

build:
	pushd cmd/icarus/; glide build; popd
	pushd cmd/icontent/; glide build; popd

install:
	pushd cmd/icarus/; glide install; popd
	pushd cmd/icontent/; glide install; popd

sys-install:
	mv cmd/icarus/icarus /usr/local/bin/
	mv cmd/icontent/icontent /usr/local/bin/

brew-redis:
	redis-server /usr/local/etc/redis.conf
