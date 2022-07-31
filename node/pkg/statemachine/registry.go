// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package statemachine

import (
	"sync"
)

func RegisterPrimitiveType[I, O any](registry *PrimitiveTypeRegistry) func(primitiveType PrimitiveType[I, O]) {
	return func(primitiveType PrimitiveType[I, O]) {
		registry.register(primitiveType.Service(), func(manager *primitiveManager, info primitiveInfo) primitiveExecutor {
			context := newPrimitiveContext[I, O](manager, info, primitiveType)
			primitive := primitiveType.NewPrimitive(context)
			return newPrimitiveStateMachineExecutor[I, O](context, primitive)
		})
	}
}

func NewPrimitiveTypeRegistry() *PrimitiveTypeRegistry {
	return &PrimitiveTypeRegistry{
		types: make(map[string]func(*primitiveManager, primitiveInfo) primitiveExecutor),
	}
}

type PrimitiveTypeRegistry struct {
	types map[string]func(*primitiveManager, primitiveInfo) primitiveExecutor
	mu    sync.RWMutex
}

func (r *PrimitiveTypeRegistry) register(service string, factory func(*primitiveManager, primitiveInfo) primitiveExecutor) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.types[service] = factory
}

func (r *PrimitiveTypeRegistry) lookup(service string) (func(*primitiveManager, primitiveInfo) primitiveExecutor, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	factory, ok := r.types[service]
	return factory, ok
}
