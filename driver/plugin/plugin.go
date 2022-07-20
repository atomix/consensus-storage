// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/atomix/multi-raft-storage/driver"
	"github.com/atomix/runtime/sdk/pkg/runtime"
)

var Plugin = driver.New(runtime.NewNetwork())
