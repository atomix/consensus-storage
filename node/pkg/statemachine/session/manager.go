// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/snapshot"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"time"
)

func NewManager(ctx statemachine.SessionManagerContext, factory NewPrimitiveManagerFunc) statemachine.SessionManager {
	sm := &sessionManagerStateMachine{
		SessionManagerContext: ctx,
		sessions:              newManagedSessions(),
		proposals:             newPrimitiveProposals(),
	}
	sm.sm = factory(sm)
	return sm
}

type sessionManagerStateMachine struct {
	statemachine.SessionManagerContext
	sm        PrimitiveManager
	sessions  *managedSessions
	proposals *primitiveProposals
	prevTime  time.Time
}

func (m *sessionManagerStateMachine) Sessions() Sessions {
	return m.sessions
}

func (m *sessionManagerStateMachine) Proposals() Proposals {
	return m.proposals
}

func (m *sessionManagerStateMachine) Snapshot(writer *snapshot.Writer) error {
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

func (m *sessionManagerStateMachine) Recover(reader *snapshot.Reader) error {
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

func (m *sessionManagerStateMachine) OpenSession(proposal statemachine.OpenSessionProposal) {
	session := newManagedSession(m)
	session.open(proposal)
	m.prevTime = m.Time()
}

func (m *sessionManagerStateMachine) KeepAlive(proposal statemachine.KeepAliveProposal) {
	sessionID := ID(proposal.Input().SessionID)
	session, ok := m.sessions.get(sessionID)
	if !ok {
		proposal.Error(errors.NewFault("session not found"))
		proposal.Close()
		return
	}
	session.keepAlive(proposal)

	// Compute the minimum session timeout
	var minSessionTimeout time.Duration
	for _, session := range m.sessions.list() {
		if session.timeout > minSessionTimeout {
			minSessionTimeout = session.timeout
		}
	}

	// Compute the maximum time at which sessions may be expired.
	// If no keep-alive has been received from any session for more than the minimum session
	// timeout, suspect a stop-the-world pause may have occurred. We decline to expire any
	// of the sessions in this scenario, instead resetting the timestamps for all the sessions.
	// Only expire a session if keep-alives have been received from other sessions during the
	// session's expiration period.
	maxExpireTime := m.prevTime.Add(minSessionTimeout)
	for _, session := range m.sessions.list() {
		session.checkExpiration(maxExpireTime)
	}
	m.prevTime = m.Time()
}

func (m *sessionManagerStateMachine) CloseSession(proposal statemachine.CloseSessionProposal) {
	sessionID := ID(proposal.Input().SessionID)
	session, ok := m.sessions.get(sessionID)
	if !ok {
		proposal.Error(errors.NewFault("session not found"))
		proposal.Close()
		return
	}
	session.close(proposal)
}

func (m *sessionManagerStateMachine) Propose(proposal statemachine.SessionProposal) {
	sessionID := ID(proposal.Input().SessionID)
	session, ok := m.sessions.get(sessionID)
	if !ok {
		proposal.Error(errors.NewFault("session not found"))
		proposal.Close()
		return
	}
	session.propose(proposal)
}

func (m *sessionManagerStateMachine) Query(query statemachine.SessionQuery) {
	sessionID := ID(query.Input().SessionID)
	session, ok := m.sessions.get(sessionID)
	if !ok {
		query.Error(errors.NewFault("session not found"))
		query.Close()
		return
	}
	session.query(query)
}
