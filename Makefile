# SPDX-FileCopyrightText: 2022-present Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

.PHONY: build
build: node controller driver

.PHONY: node
node:
	$(MAKE) -C node build

.PHONY: controller
controller:
	$(MAKE) -C controller build

.PHONY: driver
driver:
	$(MAKE) -C driver build
