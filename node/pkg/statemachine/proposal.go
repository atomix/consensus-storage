// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package statemachine

type ProposalID uint64

// Proposal is a proposal operation
type Proposal[I, O any] interface {
	Call[ProposalID, I, O]
}

func newTranscodingProposal[I1, O1, I2, O2 any](proposal Proposal[I1, O1], input I2, transcoder func(O2) O1) Proposal[I2, O2] {
	return &transcodingProposal[I1, O1, I2, O2]{
		Call: newTranscodingCall[ProposalID, I1, O1, I2, O2](proposal, input, transcoder),
	}
}

type transcodingProposal[I1, O1, I2, O2 any] struct {
	Call[ProposalID, I2, O2]
}
