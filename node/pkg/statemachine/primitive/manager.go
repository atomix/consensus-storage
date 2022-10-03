// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package primitive

import (
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/session"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/snapshot"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/atomix/runtime/sdk/pkg/logging"
)

var log = logging.GetLogger()

func NewManager(ctx session.Context, registry *TypeRegistry) session.PrimitiveManager {
	return &primitiveManager{
		Context:    ctx,
		registry:   registry,
		primitives: make(map[ID]managedPrimitive),
	}
}

type primitiveManager struct {
	session.Context
	registry   *TypeRegistry
	primitives map[ID]managedPrimitive
}

func (m *primitiveManager) Snapshot(writer *snapshot.Writer) error {
	if err := writer.WriteVarInt(len(m.primitives)); err != nil {
		return err
	}
	for _, primitive := range m.primitives {
		snapshot := &multiraftv1.PrimitiveSnapshot{
			PrimitiveID: multiraftv1.PrimitiveID(primitive.ID()),
			Spec: multiraftv1.PrimitiveSpec{
				Service:   primitive.Service(),
				Namespace: primitive.Namespace(),
				Name:      primitive.Name(),
			},
		}
		if err := writer.WriteMessage(snapshot); err != nil {
			return err
		}
		if err := primitive.Snapshot(writer); err != nil {
			return err
		}
	}
	return nil
}

func (m *primitiveManager) Recover(reader *snapshot.Reader) error {
	n, err := reader.ReadVarInt()
	if err != nil {
		return err
	}
	for i := 0; i < n; i++ {
		snapshot := &multiraftv1.PrimitiveSnapshot{}
		if err := reader.ReadMessage(snapshot); err != nil {
			return err
		}
		factory, ok := m.registry.lookup(snapshot.Spec.Service)
		if !ok {
			return errors.NewFault("primitive type not found")
		}
		primitiveID := ID(snapshot.PrimitiveID)
		primitive := factory(m.Context, primitiveID, snapshot.Spec.Namespace, snapshot.Spec.Name)
		m.primitives[primitiveID] = primitive
		if err := primitive.Recover(reader); err != nil {
			return err
		}
	}
	return nil
}

func (m *primitiveManager) Propose(proposal session.PrimitiveProposal) {
	primitive, ok := m.primitives[ID(proposal.Input().PrimitiveID)]
	if !ok {
		proposal.Error(errors.NewForbidden("primitive %d not found", proposal.Input().PrimitiveID))
		proposal.Close()
	} else {
		primitive.propose(proposal)
	}
}

func (m *primitiveManager) CreatePrimitive(proposal session.CreatePrimitiveProposal) {
	var primitive managedPrimitive
	for _, p := range m.primitives {
		if p.Namespace() == proposal.Input().Namespace &&
			p.Name() == proposal.Input().Name {
			if p.Service() != proposal.Input().Service {
				proposal.Error(errors.NewForbidden("cannot create primitive of a different type with the same name"))
				proposal.Close()
				return
			}
			primitive = p
			break
		}
	}

	if primitive == nil {
		factory, ok := m.registry.lookup(proposal.Input().Service)
		if !ok {
			proposal.Error(errors.NewForbidden("unknown primitive type"))
			proposal.Close()
			return
		} else {
			primitiveID := ID(proposal.ID())
			primitive = factory(m.Context, primitiveID, proposal.Input().Namespace, proposal.Input().Name)
			m.primitives[primitiveID] = primitive
		}
	}

	primitive.open(proposal)
}

func (m *primitiveManager) ClosePrimitive(proposal session.ClosePrimitiveProposal) {
	primitive, ok := m.primitives[ID(proposal.Input().PrimitiveID)]
	if !ok {
		proposal.Error(errors.NewForbidden("primitive %d not found", proposal.Input().PrimitiveID))
		proposal.Close()
	} else {
		primitive.close(proposal)
	}
}

func (m *primitiveManager) Query(query session.PrimitiveQuery) {
	primitive, ok := m.primitives[ID(query.Input().PrimitiveID)]
	if !ok {
		query.Error(errors.NewForbidden("primitive %d not found", query.Input().PrimitiveID))
		query.Close()
	} else {
		primitive.query(query)
	}
}
