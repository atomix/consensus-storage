// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package primitives

import "github.com/atomix/multi-raft-storage/node/pkg/statemachine"

func RegisterPrimitiveTypes(registry *statemachine.PrimitiveTypeRegistry) {
	RegisterCounterType(registry)
	RegisterMapType(registry)
}
