# SPDX-FileCopyrightText: 2022-present Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

ifdef TAG
MAJOR := $(word 1, $(subst ., , $(TAG)))
MINOR := $(word 2, $(subst ., , $(TAG)))
PATCH := $(word 3, $(subst ., , $(TAG)))
endif

.PHONY: build
build:
	docker build . -t atomix/multi-raft-node:latest

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

reuse-tool: # @HELP install reuse if not present
	command -v reuse || python3 -m pip install reuse

license: reuse-tool # @HELP run license checks
	reuse lint
