// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package statemachine

import (
	"sync"
)

func RegisterPrimitiveType[I, O any](registry *PrimitiveTypeRegistry) func(primitiveType PrimitiveType[I, O]) {
	return func(primitiveType PrimitiveType[I, O]) {
		registry.register(primitiveType.Name(), primitiveType.APIVersion(), func(manager *primitiveManager, info primitiveInfo) primitiveExecutor {
			context := newPrimitiveContext[I, O](manager, info, primitiveType)
			primitive := primitiveType.NewPrimitive(context)
			return newPrimitiveStateMachineExecutor[I, O](context, primitive)
		})
	}
}

func NewPrimitiveTypeRegistry() *PrimitiveTypeRegistry {
	return &PrimitiveTypeRegistry{
		types: make(map[registryKey]func(*primitiveManager, primitiveInfo) primitiveExecutor),
	}
}

type PrimitiveTypeRegistry struct {
	types map[registryKey]func(*primitiveManager, primitiveInfo) primitiveExecutor
	mu    sync.RWMutex
}

func (r *PrimitiveTypeRegistry) register(name, apiVersion string, factory func(*primitiveManager, primitiveInfo) primitiveExecutor) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := registryKey{
		name:       name,
		apiVersion: apiVersion,
	}
	r.types[key] = factory
}

func (r *PrimitiveTypeRegistry) lookup(name, apiVersion string) (func(*primitiveManager, primitiveInfo) primitiveExecutor, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	key := registryKey{
		name:       name,
		apiVersion: apiVersion,
	}
	factory, ok := r.types[key]
	return factory, ok
}

type registryKey struct {
	name       string
	apiVersion string
}
