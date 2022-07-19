// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package primitive

import "google.golang.org/grpc"

type Type interface {
	Name() string
	APIVersion() string
	NewStateMachine(ctx Context) StateMachine
	RegisterServices(server *grpc.Server, protocol Protocol)
}

func NewType(name, apiVersion string, registrar func(*grpc.Server, Protocol), factory func(Context) StateMachine) Type {
	return &primitiveType{
		name:       name,
		apiVersion: apiVersion,
		registrar:  registrar,
		factory:    factory,
	}
}

type primitiveType struct {
	name       string
	apiVersion string
	registrar  func(*grpc.Server, Protocol)
	factory    func(Context) StateMachine
}

func (t *primitiveType) Name() string {
	return t.name
}

func (t *primitiveType) APIVersion() string {
	return t.apiVersion
}

func (t *primitiveType) NewStateMachine(ctx Context) StateMachine {
	return t.factory(ctx)
}

func (t *primitiveType) RegisterServices(server *grpc.Server, protocol Protocol) {
	t.registrar(server, protocol)
}
