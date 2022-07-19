// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package primitive

import "sync"

type Registry struct {
	types map[registryKey]Type
	mu    sync.RWMutex
}

func (r *Registry) Register(t Type) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := registryKey{
		name:       t.Name(),
		apiVersion: t.APIVersion(),
	}
	r.types[key] = t
}

func (r *Registry) Get(name, apiVersion string) Type {
	r.mu.RLock()
	defer r.mu.RUnlock()
	key := registryKey{
		name:       name,
		apiVersion: apiVersion,
	}
	return r.types[key]
}

type registryKey struct {
	name       string
	apiVersion string
}
