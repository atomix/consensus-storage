// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package primitive

import (
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/session"
	"github.com/atomix/runtime/sdk/pkg/errors"
	"github.com/atomix/runtime/sdk/pkg/logging"
	"github.com/google/uuid"
	"time"
)

type ProposalID = session.ProposalID

type ProposalState = session.ProposalState

// Proposal is a proposal operation
type Proposal[I, O any] session.Proposal[I, O]

// Proposals provides access to pending proposals
type Proposals[I, O any] interface {
	// Get gets a proposal by ID
	Get(id ProposalID) (Proposal[I, O], bool)
	// List lists all open proposals
	List() []Proposal[I, O]
}

func newPrimitiveProposal[I, O any](session *primitiveSession[I, O]) *primitiveProposal[I, O] {
	return &primitiveProposal[I, O]{
		session: session,
	}
}

type primitiveProposal[I, O any] struct {
	parent   session.Proposal[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput]
	session  *primitiveSession[I, O]
	input    I
	state    CallState
	watchers map[uuid.UUID]func(CallState)
	cancel   session.CancelFunc
	log      logging.Logger
}

func (p *primitiveProposal[I, O]) ID() ProposalID {
	return p.parent.ID()
}

func (p *primitiveProposal[I, O]) Log() logging.Logger {
	return p.log
}

func (p *primitiveProposal[I, O]) Time() time.Time {
	return p.parent.Time()
}

func (p *primitiveProposal[I, O]) State() ProposalState {
	return p.state
}

func (p *primitiveProposal[I, O]) Watch(watcher func(ProposalState)) CancelFunc {
	if p.state != Running {
		watcher(p.state)
		return func() {}
	}
	if p.watchers == nil {
		p.watchers = make(map[uuid.UUID]func(ProposalState))
	}
	id := uuid.New()
	p.watchers[id] = watcher
	return func() {
		delete(p.watchers, id)
	}
}

func (p *primitiveProposal[I, O]) Session() Session {
	return p.session
}

func (p *primitiveProposal[I, O]) init(parent session.Proposal[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput]) error {
	input, err := p.session.primitive.codec.DecodeInput(parent.Input().Payload)
	if err != nil {
		return err
	}

	p.parent = parent
	p.input = input
	p.state = Running
	p.log = p.session.Log().WithFields(logging.Uint64("ProposalID", uint64(parent.ID())))
	p.cancel = parent.Watch(func(state session.ProposalState) {
		if p.state != Running {
			return
		}
		switch state {
		case session.Complete:
			p.destroy(Complete)
		case session.Canceled:
			p.destroy(Canceled)
		}
	})
	p.session.primitive.proposals.add(p)
	p.session.proposals[p.ID()] = p
	return nil
}

func (p *primitiveProposal[I, O]) execute(parent session.Proposal[*multiraftv1.PrimitiveProposalInput, *multiraftv1.PrimitiveProposalOutput]) {
	if p.state != Pending {
		return
	}

	if err := p.init(parent); err != nil {
		p.Log().Errorw("Failed decoding proposal", logging.Error("Error", err))
		parent.Error(errors.NewInternal("failed decoding proposal: %s", err.Error()))
		parent.Close()
	} else {
		p.session.primitive.sm.Propose(p)
		if p.state == Running {
			p.session.registerProposal(p)
		}
	}
}

func (p *primitiveProposal[I, O]) Input() I {
	return p.input
}

func (p *primitiveProposal[I, O]) Output(output O) {
	if p.state != Running {
		return
	}
	payload, err := p.session.primitive.codec.EncodeOutput(output)
	if err != nil {
		p.Log().Errorw("Failed encoding proposal", logging.Error("Error", err))
		p.parent.Error(errors.NewInternal("failed encoding proposal: %s", err.Error()))
		p.parent.Close()
	} else {
		p.parent.Output(&multiraftv1.PrimitiveProposalOutput{
			Payload: payload,
		})
	}
}

func (p *primitiveProposal[I, O]) Error(err error) {
	if p.state != Running {
		return
	}
	p.parent.Error(err)
	p.parent.Close()
}

func (p *primitiveProposal[I, O]) Cancel() {
	if p.state != Running {
		return
	}
	p.parent.Cancel()
}

func (p *primitiveProposal[I, O]) Close() {
	if p.state != Running {
		return
	}
	p.parent.Close()
}

func (p *primitiveProposal[I, O]) destroy(state ProposalState) {
	p.cancel()
	p.state = state
	p.session.primitive.proposals.remove(p.ID())
	p.session.unregisterProposal(p.ID())
	if p.watchers != nil {
		for _, watcher := range p.watchers {
			watcher(state)
		}
	}
}

var _ Proposal[any, any] = (*primitiveProposal[any, any])(nil)

func newPrimitiveProposals[I, O any]() *primitiveProposals[I, O] {
	return &primitiveProposals[I, O]{
		proposals: make(map[statemachine.ProposalID]*primitiveProposal[I, O]),
	}
}

type primitiveProposals[I, O any] struct {
	proposals map[statemachine.ProposalID]*primitiveProposal[I, O]
}

func (p *primitiveProposals[I, O]) Get(id statemachine.ProposalID) (Proposal[I, O], bool) {
	proposal, ok := p.proposals[id]
	return proposal, ok
}

func (p *primitiveProposals[I, O]) List() []Proposal[I, O] {
	proposals := make([]Proposal[I, O], 0, len(p.proposals))
	for _, proposal := range p.proposals {
		proposals = append(proposals, proposal)
	}
	return proposals
}

func (p *primitiveProposals[I, O]) add(proposal *primitiveProposal[I, O]) {
	p.proposals[proposal.ID()] = proposal
}

func (p *primitiveProposals[I, O]) remove(id statemachine.ProposalID) {
	delete(p.proposals, id)
}

var _ Proposals[any, any] = (*primitiveProposals[any, any])(nil)
