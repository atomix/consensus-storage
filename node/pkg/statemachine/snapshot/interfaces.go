// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package snapshot

type Recoverable interface {
	Snapshot(writer *Writer) error
	Recover(reader *Reader) error
}
