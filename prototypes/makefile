all:

build-docker:
	docker build -t file-sync-protocol:1 --build-arg prototype=proto-1 .
	docker build -t file-sync-protocol:2b --build-arg prototype=proto-2b .
	docker build -t file-sync-protocol:3 --build-arg prototype=proto-3 .

clean:
	docker rmi --force proto-1 proto-2b proto-3
