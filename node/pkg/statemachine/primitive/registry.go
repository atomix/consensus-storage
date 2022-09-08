// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package primitive

import (
	"sync"
)

func RegisterType[I, O any](registry *TypeRegistry) func(primitiveType Type[I, O]) {
	return func(primitiveType Type[I, O]) {
		registry.register(primitiveType.Service(), func(context *primitiveContext) primitiveDelegate {
			return newGenericPrimitive[I, O](context, primitiveType)
		})
	}
}

func NewTypeRegistry() *TypeRegistry {
	return &TypeRegistry{
		types: make(map[string]func(*primitiveContext) primitiveDelegate),
	}
}

type TypeRegistry struct {
	types map[string]func(*primitiveContext) primitiveDelegate
	mu    sync.RWMutex
}

func (r *TypeRegistry) register(service string, factory func(*primitiveContext) primitiveDelegate) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.types[service] = factory
}

func (r *TypeRegistry) lookup(service string) (func(*primitiveContext) primitiveDelegate, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	factory, ok := r.types[service]
	return factory, ok
}
