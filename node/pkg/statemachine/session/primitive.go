// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/snapshot"
)

type NewPrimitiveManagerFunc func(Context) PrimitiveManager

type PrimitiveManager interface {
	snapshot.Recoverable
	CreatePrimitive(proposal CreatePrimitiveProposal)
	ClosePrimitive(proposal ClosePrimitiveProposal)
	Propose(proposal PrimitiveProposal)
	Query(query PrimitiveQuery)
}

type CreatePrimitiveProposal Proposal[*multiraftv1.CreatePrimitiveInput, *multiraftv1.CreatePrimitiveOutput]
type ClosePrimitiveProposal Proposal[*multiraftv1.ClosePrimitiveInput, *multiraftv1.ClosePrimitiveOutput]
type PrimitiveProposal Proposal[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput]
type PrimitiveQuery Query[*multiraftv1.PrimitiveQueryInput, *multiraftv1.PrimitiveQueryOutput]

type Context interface {
	statemachine.SessionManagerContext
	// Sessions returns the open sessions
	Sessions() Sessions
	// Proposals returns the pending proposals
	Proposals() Proposals
}
