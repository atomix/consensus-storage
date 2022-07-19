// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package snapshot

import (
	"github.com/gogo/protobuf/proto"
	"io"
)

func NewReader(reader io.Reader) *Reader {
	return &Reader{
		Reader: reader,
	}
}

type Reader struct {
	io.Reader
}

// ReadBool reads a boolean from the given reader
func (r *Reader) ReadBool() (bool, error) {
	bytes := make([]byte, 1)
	if _, err := r.Read(bytes); err != nil {
		return false, err
	}
	return bytes[0] == 1, nil
}

// ReadVarUint64 reads an unsigned variable length integer from the given reader
func (r *Reader) ReadVarUint64() (uint64, error) {
	var x uint64
	var s uint
	bytes := make([]byte, 1)
	for i := 0; i <= 9; i++ {
		if n, err := r.Read(bytes); err != nil || n == -1 {
			return 0, err
		}
		b := bytes[0]
		if b < 0x80 {
			if i == 9 && b > 1 {
				return 0, nil
			}
			return x | uint64(b)<<s, nil
		}
		x |= uint64(b&0x7f) << s
		s += 7
	}
	return 0, nil
}

// ReadVarInt64 reads a signed variable length integer from the given reader
func (r *Reader) ReadVarInt64() (int64, error) {
	ux, n := r.ReadVarUint64()
	x := int64(ux >> 1)
	if ux&1 != 0 {
		x = ^x
	}
	return x, n
}

// ReadVarInt32 reads a signed 32-bit integer from the given reader
func (r *Reader) ReadVarInt32() (int32, error) {
	i, err := r.ReadVarInt64()
	if err != nil {
		return 0, err
	}
	return int32(i), nil
}

// ReadVarInt reads a signed integer from the given reader
func (r *Reader) ReadVarInt() (int, error) {
	i, err := r.ReadVarInt64()
	if err != nil {
		return 0, err
	}
	return int(i), nil
}

// ReadVarUint32 reads an unsigned 32-bit integer from the given reader
func (r *Reader) ReadVarUint32() (uint32, error) {
	i, err := r.ReadVarUint64()
	if err != nil {
		return 0, err
	}
	return uint32(i), nil
}

// ReadVarUint reads an unsigned integer from the given reader
func (r *Reader) ReadVarUint() (uint, error) {
	i, err := r.ReadVarUint64()
	if err != nil {
		return 0, err
	}
	return uint(i), nil
}

// ReadBytes reads a byte slice from the given reader
func (r *Reader) ReadBytes() ([]byte, error) {
	length, err := r.ReadVarInt()
	if err != nil {
		return nil, err
	}
	if length == 0 {
		return []byte{}, nil
	}
	bytes := make([]byte, length)
	if _, err := r.Read(bytes); err != nil {
		return nil, err
	}
	return bytes, nil
}

func (r *Reader) ReadMessage(message proto.Message) error {
	bytes, err := r.ReadBytes()
	if err != nil {
		return err
	}
	return proto.Unmarshal(bytes, message)
}
