// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package primitive

import (
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/gogo/protobuf/proto"
	"sync"
)

func RegisterType[I, O proto.Message](registry *TypeRegistry) func(primitiveType Type[I, O]) {
	return func(primitiveType Type[I, O]) {
		registry.register(primitiveType.Service(), func(context Context[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput]) primitiveDelegate {
			return newPrimitiveDelegate[I, O](context, primitiveType)
		})
	}
}

func NewTypeRegistry() *TypeRegistry {
	return &TypeRegistry{
		types: make(map[string]func(Context[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput]) primitiveDelegate),
	}
}

type TypeRegistry struct {
	types map[string]func(Context[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput]) primitiveDelegate
	mu    sync.RWMutex
}

func (r *TypeRegistry) register(service string, factory func(Context[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput]) primitiveDelegate) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.types[service] = factory
}

func (r *TypeRegistry) lookup(service string) (func(Context[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput]) primitiveDelegate, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	factory, ok := r.types[service]
	return factory, ok
}
