// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package primitive

import (
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/session"
	"sync"
)

func RegisterType[I, O any](registry *TypeRegistry) func(primitiveType Type[I, O]) {
	return func(primitiveType Type[I, O]) {
		registry.register(primitiveType.Service(), func(context session.Context, id ID, spec multiraftv1.PrimitiveSpec) managedPrimitive {
			return newPrimitive[I, O](context, id, spec, primitiveType)
		})
	}
}

func NewTypeRegistry() *TypeRegistry {
	return &TypeRegistry{
		types: make(map[string]func(session.Context, ID, multiraftv1.PrimitiveSpec) managedPrimitive),
	}
}

type TypeRegistry struct {
	types map[string]func(session.Context, ID, multiraftv1.PrimitiveSpec) managedPrimitive
	mu    sync.RWMutex
}

func (r *TypeRegistry) register(service string, factory func(session.Context, ID, multiraftv1.PrimitiveSpec) managedPrimitive) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.types[service] = factory
}

func (r *TypeRegistry) lookup(service string) (func(session.Context, ID, multiraftv1.PrimitiveSpec) managedPrimitive, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	factory, ok := r.types[service]
	return factory, ok
}
