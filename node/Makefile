# SPDX-FileCopyrightText: 2022-present Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

ifdef TAG
VERSION := $(word 2, $(subst /, , $(TAG)))
MAJOR := $(word 1, $(subst ., , $(VERSION)))
MINOR := $(word 2, $(subst ., , $(VERSION)))
PATCH := $(word 3, $(subst ., , $(VERSION)))
endif

.PHONY: build
build: test
	docker build . -t atomix/multi-raft-node:latest

.PHONY: test
test:
	CGO_ENABLED=1 go test -race github.com/atomix/multi-raft-storage/node/...

.PHONY: release
release: build
	docker push atomix/multi-raft-node:latest
ifdef TAG
	docker tag atomix/multi-raft-node:latest atomix/multi-raft-node:${MAJOR}.${MINOR}.${PATCH}
	docker tag atomix/multi-raft-node:latest atomix/multi-raft-node:${MAJOR}.${MINOR}
	docker push atomix/multi-raft-node:${MAJOR}.${MINOR}.${PATCH}
	docker push atomix/multi-raft-node:${MAJOR}.${MINOR}
endif

kind: build
	@if [ "`kind get clusters`" = '' ]; then echo "no kind cluster found" && exit 1; fi
	kind load docker-image atomix/multi-raft-node:latest

generate:
	@find . -name '*.pb.go' -delete
	docker run -i \
		-v `pwd`:/build \
		atomix/codegen:go-latest \
	    --proto-path ./pkg --go-path ./pkg --import-path github.com/atomix/multi-raft-storage/node/pkg

reuse-tool: # @HELP install reuse if not present
	command -v reuse || python3 -m pip install reuse

license: reuse-tool # @HELP run license checks
	reuse lint
