// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/snapshot"
	"github.com/atomix/runtime/sdk/pkg/errors"
)

func NewManager(ctx statemachine.SessionManagerContext, factory NewPrimitiveManagerFunc) statemachine.SessionManager {
	sm := &sessionManager{
		SessionManagerContext: ctx,
		sessions:              newManagedSessions(),
		proposals:             newPrimitiveProposals(),
	}
	sm.sm = factory(sm)
	return sm
}

type sessionManager struct {
	statemachine.SessionManagerContext
	sm        PrimitiveManager
	sessions  *managedSessions
	proposals *primitiveProposals
}

func (m *sessionManager) Sessions() Sessions {
	return m.sessions
}

func (m *sessionManager) Proposals() Proposals {
	return m.proposals
}

func (m *sessionManager) Snapshot(writer *snapshot.Writer) error {
	sessions := m.sessions.list()
	if err := writer.WriteVarInt(len(sessions)); err != nil {
		return err
	}
	for _, session := range sessions {
		if err := session.Snapshot(writer); err != nil {
			return err
		}
	}
	return m.sm.Snapshot(writer)
}

func (m *sessionManager) Recover(reader *snapshot.Reader) error {
	n, err := reader.ReadVarInt()
	if err != nil {
		return err
	}
	for i := 0; i < n; i++ {
		session := newManagedSession(m)
		if err := session.Recover(reader); err != nil {
			return err
		}
	}
	return m.sm.Recover(reader)
}

func (m *sessionManager) OpenSession(openSession statemachine.OpenSessionProposal) {
	session := newManagedSession(m)
	session.open(openSession)
}

func (m *sessionManager) KeepAlive(keepAlive statemachine.KeepAliveProposal) {
	sessionID := ID(keepAlive.Input().SessionID)
	session, ok := m.sessions.get(sessionID)
	if !ok {
		keepAlive.Error(errors.NewForbidden("session not found"))
		keepAlive.Close()
	} else {
		session.keepAlive(keepAlive)
	}
}

func (m *sessionManager) CloseSession(closeSession statemachine.CloseSessionProposal) {
	sessionID := ID(closeSession.Input().SessionID)
	session, ok := m.sessions.get(sessionID)
	if !ok {
		closeSession.Error(errors.NewForbidden("session not found"))
		closeSession.Close()
	} else {
		session.close(closeSession)
	}
}

func (m *sessionManager) Propose(proposal statemachine.SessionProposal) {
	sessionID := ID(proposal.Input().SessionID)
	session, ok := m.sessions.get(sessionID)
	if !ok {
		proposal.Error(errors.NewForbidden("session not found"))
		proposal.Close()
	} else {
		session.propose(proposal)
	}
}

func (m *sessionManager) Query(query statemachine.SessionQuery) {
	sessionID := ID(query.Input().SessionID)
	session, ok := m.sessions.get(sessionID)
	if !ok {
		query.Error(errors.NewForbidden("session not found"))
		query.Close()
	} else {
		session.query(query)
	}
}
