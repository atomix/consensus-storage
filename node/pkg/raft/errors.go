// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package raft

import (
	multiraftv1 "github.com/atomix/multi-raft/api/atomix/multiraft/v1"
	"github.com/atomix/runtime/pkg/errors"
)

// getErrorFromStatus creates a typed error from a response status
func getErrorFromStatus(status multiraftv1.OperationOutput_Status, message string) error {
	switch status {
	case multiraftv1.OperationOutput_OK:
		return nil
	case multiraftv1.OperationOutput_ERROR:
		return errors.NewUnknown(message)
	case multiraftv1.OperationOutput_UNKNOWN:
		return errors.NewUnknown(message)
	case multiraftv1.OperationOutput_CANCELED:
		return errors.NewCanceled(message)
	case multiraftv1.OperationOutput_NOT_FOUND:
		return errors.NewNotFound(message)
	case multiraftv1.OperationOutput_ALREADY_EXISTS:
		return errors.NewAlreadyExists(message)
	case multiraftv1.OperationOutput_UNAUTHORIZED:
		return errors.NewUnauthorized(message)
	case multiraftv1.OperationOutput_FORBIDDEN:
		return errors.NewForbidden(message)
	case multiraftv1.OperationOutput_CONFLICT:
		return errors.NewConflict(message)
	case multiraftv1.OperationOutput_INVALID:
		return errors.NewInvalid(message)
	case multiraftv1.OperationOutput_UNAVAILABLE:
		return errors.NewUnavailable(message)
	case multiraftv1.OperationOutput_NOT_SUPPORTED:
		return errors.NewNotSupported(message)
	case multiraftv1.OperationOutput_TIMEOUT:
		return errors.NewTimeout(message)
	case multiraftv1.OperationOutput_FAULT:
		return errors.NewFault(message)
	case multiraftv1.OperationOutput_INTERNAL:
		return errors.NewInternal(message)
	default:
		return errors.NewUnknown(message)
	}
}

// getStatus gets the proto status for the given error
func getStatus(err error) multiraftv1.OperationOutput_Status {
	if err == nil {
		return multiraftv1.OperationOutput_OK
	}

	typed, ok := err.(*errors.TypedError)
	if !ok {
		return multiraftv1.OperationOutput_ERROR
	}

	switch typed.Type {
	case errors.Unknown:
		return multiraftv1.OperationOutput_UNKNOWN
	case errors.Canceled:
		return multiraftv1.OperationOutput_CANCELED
	case errors.NotFound:
		return multiraftv1.OperationOutput_NOT_FOUND
	case errors.AlreadyExists:
		return multiraftv1.OperationOutput_ALREADY_EXISTS
	case errors.Unauthorized:
		return multiraftv1.OperationOutput_UNAUTHORIZED
	case errors.Forbidden:
		return multiraftv1.OperationOutput_FORBIDDEN
	case errors.Conflict:
		return multiraftv1.OperationOutput_CONFLICT
	case errors.Invalid:
		return multiraftv1.OperationOutput_INVALID
	case errors.Unavailable:
		return multiraftv1.OperationOutput_UNAVAILABLE
	case errors.NotSupported:
		return multiraftv1.OperationOutput_NOT_SUPPORTED
	case errors.Timeout:
		return multiraftv1.OperationOutput_TIMEOUT
	case errors.Fault:
		return multiraftv1.OperationOutput_FAULT
	case errors.Internal:
		return multiraftv1.OperationOutput_INTERNAL
	default:
		return multiraftv1.OperationOutput_ERROR
	}
}

// getMessage gets the message for the given error
func getMessage(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
