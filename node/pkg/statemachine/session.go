// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package statemachine

import (
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/snapshot"
	"time"
)

type NewSessionManagerFunc func(SessionManagerContext) SessionManager

type SessionManager interface {
	snapshot.Recoverable
	OpenSession(proposal OpenSessionProposal)
	KeepAlive(proposal KeepAliveProposal)
	CloseSession(proposal CloseSessionProposal)
	Propose(proposal SessionProposal)
	Query(query SessionQuery)
}

type OpenSessionProposal Proposal[*multiraftv1.OpenSessionInput, *multiraftv1.OpenSessionOutput]
type KeepAliveProposal Proposal[*multiraftv1.KeepAliveInput, *multiraftv1.KeepAliveOutput]
type CloseSessionProposal Proposal[*multiraftv1.CloseSessionInput, *multiraftv1.CloseSessionOutput]
type SessionProposal Proposal[*multiraftv1.SessionProposalInput, *multiraftv1.SessionProposalOutput]
type SessionQuery Query[*multiraftv1.SessionQueryInput, *multiraftv1.SessionQueryOutput]

type SessionManagerContext interface {
	Context
	// Time returns the current service time
	Time() time.Time
	// Scheduler returns the service scheduler
	Scheduler() Scheduler
}

type CancelFunc func()
