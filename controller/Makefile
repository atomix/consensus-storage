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
build:
	docker build . -t atomix/consensus-controller:latest -f build/Dockerfile

.PHONY: release
release: build
	docker push atomix/consensus-controller:latest
ifdef TAG
	docker tag atomix/consensus-controller:latest atomix/consensus-controller:${MAJOR}.${MINOR}.${PATCH}
	docker tag atomix/consensus-controller:latest atomix/consensus-controller:${MAJOR}.${MINOR}
	docker push atomix/consensus-controller:${MAJOR}.${MINOR}.${PATCH}
	docker push atomix/consensus-controller:${MAJOR}.${MINOR}
endif

kind: build
	@if [ "`kind get clusters`" = '' ]; then echo "no kind cluster found" && exit 1; fi
	kind load docker-image atomix/consensus-controller:latest

reuse-tool: # @HELP install reuse if not present
	command -v reuse || python3 -m pip install reuse

license: reuse-tool # @HELP run license checks
	reuse lint
