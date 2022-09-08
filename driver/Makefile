# SPDX-FileCopyrightText: 2022-present Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

.PHONY: build
build:
	go build ./pkg/...

.PHONY: test
test:
	CGO_ENABLED=1 go test -race github.com/atomix/multi-raft-storage/driver/...

reuse-tool: # @HELP install reuse if not present
	command -v reuse || python3 -m pip install reuse

license: reuse-tool # @HELP run license checks
	reuse lint
