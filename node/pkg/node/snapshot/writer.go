// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package snapshot

import (
	"github.com/gogo/protobuf/proto"
	"io"
)

func NewWriter(writer io.Writer) *Writer {
	return &Writer{
		Writer: writer,
	}
}

type Writer struct {
	io.Writer
}

// WriteBool writes a boolean to the given writer
func (w *Writer) WriteBool(b bool) error {
	if b {
		if _, err := w.Write([]byte{1}); err != nil {
			return err
		}
	} else {
		if _, err := w.Write([]byte{0}); err != nil {
			return err
		}
	}
	return nil
}

// WriteVarUint64 writes an unsigned variable length integer to the given writer
func (w *Writer) WriteVarUint64(x uint64) error {
	for x >= 0x80 {
		if _, err := w.Write([]byte{byte(x) | 0x80}); err != nil {
			return err
		}
		x >>= 7
	}
	if _, err := w.Write([]byte{byte(x)}); err != nil {
		return err
	}
	return nil
}

// WriteVarInt64 writes a signed variable length integer to the given writer
func (w *Writer) WriteVarInt64(x int64) error {
	ux := uint64(x) << 1
	if x < 0 {
		ux = ^ux
	}
	return w.WriteVarUint64(ux)
}

// WriteVarInt32 writes a signed 32-bit integer to the given writer
func (w *Writer) WriteVarInt32(i int32) error {
	return w.WriteVarInt64(int64(i))
}

// WriteVarInt writes a signed integer to the given writer
func (w *Writer) WriteVarInt(i int) error {
	return w.WriteVarInt64(int64(i))
}

// WriteVarUint32 writes an unsigned 32-bit integer to the given writer
func (w *Writer) WriteVarUint32(i uint32) error {
	return w.WriteVarUint64(uint64(i))
}

// WriteVarUint writes an unsigned integer to the given writer
func (w *Writer) WriteVarUint(i uint) error {
	return w.WriteVarUint64(uint64(i))
}

// WriteBytes writes a byte slice to the given writer
func (w *Writer) WriteBytes(bytes []byte) error {
	if err := w.WriteVarInt(len(bytes)); err != nil {
		return err
	}
	if _, err := w.Write(bytes); err != nil {
		return err
	}
	return nil
}

func (w *Writer) WriteMessage(message proto.Message) error {
	bytes, err := proto.Marshal(message)
	if err != nil {
		return err
	}
	return w.WriteBytes(bytes)
}
