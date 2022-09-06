// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/snapshot"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"time"
)

func NewManager(ctx statemachine.Context, factory NewPrimitiveManagerFunc) statemachine.SessionManager {
	sm := &sessionManagerStateMachine{
		Context:   ctx,
		sessions:  newManagedSessions(),
		proposals: newSessionProposals(),
	}
	sm.sm = factory(sm)
	return sm
}

type sessionManagerStateMachine struct {
	statemachine.Context
	sm        PrimitiveManager
	sessions  *managedSessions
	proposals *sessionProposals
	prevTime  time.Time
}

func (m *sessionManagerStateMachine) Sessions() Sessions[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput] {
	return m.sessions
}

func (m *sessionManagerStateMachine) Proposals() Proposals[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput] {
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

func (m *sessionManagerStateMachine) OpenSession(proposal statemachine.Proposal[*multiraftv1.OpenSessionInput, *multiraftv1.OpenSessionOutput]) {
	session := newManagedSession(m)
	session.open(proposal)
	m.prevTime = m.Time()
}

func (m *sessionManagerStateMachine) KeepAlive(proposal statemachine.Proposal[*multiraftv1.KeepAliveInput, *multiraftv1.KeepAliveOutput]) {
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

func (m *sessionManagerStateMachine) CloseSession(proposal statemachine.Proposal[*multiraftv1.CloseSessionInput, *multiraftv1.CloseSessionOutput]) {
	sessionID := ID(proposal.Input().SessionID)
	session, ok := m.sessions.get(sessionID)
	if !ok {
		proposal.Error(errors.NewFault("session not found"))
		proposal.Close()
		return
	}
	session.close(proposal)
}

func (m *sessionManagerStateMachine) Propose(parent statemachine.Proposal[*multiraftv1.SessionProposalInput, *multiraftv1.SessionProposalOutput]) {
	sessionID := ID(parent.Input().SessionID)
	session, ok := m.sessions.get(sessionID)
	if !ok {
		parent.Error(errors.NewFault("session not found"))
		parent.Close()
		return
	}

	if proposal, ok := session.proposals.get(parent.Input().SequenceNum); ok {
		proposal.replay(parent)
	} else {
		proposal := newSessionProposal(session)
		proposal.execute(parent)
	}
}

func (m *sessionManagerStateMachine) Query(query statemachine.Query[*multiraftv1.SessionQueryInput, *multiraftv1.SessionQueryOutput]) {

}
