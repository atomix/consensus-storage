// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package statemachine

import (
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/runtime/sdk/pkg/errors"
)

// getError creates a typed error from a response status
func getError(failure *multiraftv1.Failure) error {
	if failure == nil {
		return nil
	}
	switch failure.Status {
	case multiraftv1.Failure_ERROR:
		return errors.NewUnknown(failure.Message)
	case multiraftv1.Failure_UNKNOWN:
		return errors.NewUnknown(failure.Message)
	case multiraftv1.Failure_CANCELED:
		return errors.NewCanceled(failure.Message)
	case multiraftv1.Failure_NOT_FOUND:
		return errors.NewNotFound(failure.Message)
	case multiraftv1.Failure_ALREADY_EXISTS:
		return errors.NewAlreadyExists(failure.Message)
	case multiraftv1.Failure_UNAUTHORIZED:
		return errors.NewUnauthorized(failure.Message)
	case multiraftv1.Failure_FORBIDDEN:
		return errors.NewForbidden(failure.Message)
	case multiraftv1.Failure_CONFLICT:
		return errors.NewConflict(failure.Message)
	case multiraftv1.Failure_INVALID:
		return errors.NewInvalid(failure.Message)
	case multiraftv1.Failure_UNAVAILABLE:
		return errors.NewUnavailable(failure.Message)
	case multiraftv1.Failure_NOT_SUPPORTED:
		return errors.NewNotSupported(failure.Message)
	case multiraftv1.Failure_TIMEOUT:
		return errors.NewTimeout(failure.Message)
	case multiraftv1.Failure_FAULT:
		return errors.NewFault(failure.Message)
	case multiraftv1.Failure_INTERNAL:
		return errors.NewInternal(failure.Message)
	default:
		return errors.NewUnknown(failure.Message)
	}
}

// getFailure gets the proto status for the given error
func getFailure(err error) *multiraftv1.Failure {
	if err == nil {
		return nil
	}
	return &multiraftv1.Failure{
		Status:  getStatus(err),
		Message: getMessage(err),
	}
}

func getStatus(err error) multiraftv1.Failure_Status {
	typed, ok := err.(*errors.TypedError)
	if !ok {
		return multiraftv1.Failure_ERROR
	}

	switch typed.Type {
	case errors.Unknown:
		return multiraftv1.Failure_UNKNOWN
	case errors.Canceled:
		return multiraftv1.Failure_CANCELED
	case errors.NotFound:
		return multiraftv1.Failure_NOT_FOUND
	case errors.AlreadyExists:
		return multiraftv1.Failure_ALREADY_EXISTS
	case errors.Unauthorized:
		return multiraftv1.Failure_UNAUTHORIZED
	case errors.Forbidden:
		return multiraftv1.Failure_FORBIDDEN
	case errors.Conflict:
		return multiraftv1.Failure_CONFLICT
	case errors.Invalid:
		return multiraftv1.Failure_INVALID
	case errors.Unavailable:
		return multiraftv1.Failure_UNAVAILABLE
	case errors.NotSupported:
		return multiraftv1.Failure_NOT_SUPPORTED
	case errors.Timeout:
		return multiraftv1.Failure_TIMEOUT
	case errors.Fault:
		return multiraftv1.Failure_FAULT
	case errors.Internal:
		return multiraftv1.Failure_INTERNAL
	default:
		return multiraftv1.Failure_ERROR
	}
}

// getMessage gets the message for the given error
func getMessage(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
