// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package statemachine

import (
	"bytes"
	"github.com/atomix/multi-raft-storage/node/pkg/statemachine/snapshot"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestStateMachineContext(t *testing.T) {
	context := newStateMachineContext(nil)
	assert.Equal(t, time.UnixMilli(1), context.update(time.UnixMilli(1)))
	assert.Equal(t, time.UnixMilli(1), context.Time())
	assert.Equal(t, time.UnixMilli(2), context.update(time.UnixMilli(2)))
	assert.Equal(t, time.UnixMilli(2), context.Time())
	assert.Equal(t, time.UnixMilli(2), context.update(time.UnixMilli(1)))
	assert.Equal(t, time.UnixMilli(2), context.Time())
	buf := &bytes.Buffer{}
	assert.NoError(t, context.Snapshot(snapshot.NewWriter(buf)))
	context = newStateMachineContext(nil)
	assert.NoError(t, context.Recover(snapshot.NewReader(buf)))
	assert.Equal(t, time.UnixMilli(2), context.Time())
}
