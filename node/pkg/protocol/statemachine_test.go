// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	"bytes"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/snapshot"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStateMachineContext(t *testing.T) {
	context := &stateMachineContext{}
	assert.Equal(t, statemachine.Index(0), context.Index())
	assert.Equal(t, statemachine.Index(1), context.update())
	assert.Equal(t, statemachine.Index(1), context.Index())
	buf := &bytes.Buffer{}
	assert.NoError(t, context.Snapshot(snapshot.NewWriter(buf)))
	context = &stateMachineContext{}
	assert.NoError(t, context.Recover(snapshot.NewReader(buf)))
	assert.Equal(t, statemachine.Index(1), context.Index())
}
