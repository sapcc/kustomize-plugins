.PHONY: release vendor

release:
	goreleaser $@ --rm-dist

vendor:
	find . -type f -name go.mod | sed -E "s|/[^/]+$||" | xargs -L 1 bash -c 'echo "running go mod vendor in $0" && cd "$0" && go mod vendor'



  
