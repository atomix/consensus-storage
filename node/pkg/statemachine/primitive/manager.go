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
		primitives: make(map[ID]*primitiveContext),
	}
}

type primitiveManager struct {
	session.Context
	registry   *TypeRegistry
	primitives map[ID]*primitiveContext
}

func (m *primitiveManager) Snapshot(writer *snapshot.Writer) error {
	if err := writer.WriteVarInt(len(m.primitives)); err != nil {
		return err
	}
	for _, primitive := range m.primitives {
		snapshot := &multiraftv1.PrimitiveSnapshot{
			PrimitiveID: multiraftv1.PrimitiveID(primitive.ID()),
			Spec:        primitive.spec,
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
		primitive, ok := newPrimitive(m.Context, ID(snapshot.PrimitiveID), snapshot.Spec, m.registry)
		if !ok {
			return errors.NewFault("primitive type not found")
		}
		m.primitives[primitive.ID()] = primitive
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
	var primitive *primitiveContext
	for _, p := range m.primitives {
		if p.spec.Namespace == proposal.Input().Namespace &&
			p.spec.Name == proposal.Input().Name {
			if p.spec.Service != proposal.Input().Service {
				proposal.Error(errors.NewForbidden("cannot create primitive of a different type with the same name"))
				proposal.Close()
				return
			}
			primitive = p
			break
		}
	}

	if primitive == nil {
		if p, ok := newPrimitive(m.Context, ID(proposal.ID()), proposal.Input().PrimitiveSpec, m.registry); !ok {
			proposal.Error(errors.NewForbidden("unknown primitive type"))
			proposal.Close()
			return
		} else {
			m.primitives[p.ID()] = p
			primitive = p
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
