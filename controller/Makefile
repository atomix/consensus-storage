# SPDX-FileCopyrightText: 2022-present Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

VERSION ?= latest

.PHONY: build
build:
	docker build . -t atomix/multi-raft-controller:latest -f build/Dockerfile

.PHONY: release
release: build
	docker tag atomix/multi-raft-controller:latest atomix/multi-raft-controller:$(VERSION)
	docker push atomix/multi-raft-controller:latest
	docker push atomix/multi-raft-controller:$(VERSION)

kind: build
	@if [ "`kind get clusters`" = '' ]; then echo "no kind cluster found" && exit 1; fi
	kind load docker-image atomix/multi-raft-controller:latest

reuse-tool: # @HELP install reuse if not present
	command -v reuse || python3 -m pip install reuse

license: reuse-tool # @HELP run license checks
	reuse lint
