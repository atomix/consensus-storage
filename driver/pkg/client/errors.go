// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package client

import (
	multiraftv1 "github.com/atomix/multi-raft-storage/api/atomix/multiraft/v1"
	"github.com/atomix/runtime/sdk/pkg/errors"
)

// getErrorFromStatus creates a typed error from a response status
func getErrorFromStatus(status multiraftv1.OperationResponseHeaders_Status, message string) error {
	switch status {
	case multiraftv1.OperationResponseHeaders_OK:
		return nil
	case multiraftv1.OperationResponseHeaders_ERROR:
		return errors.NewUnknown(message)
	case multiraftv1.OperationResponseHeaders_UNKNOWN:
		return errors.NewUnknown(message)
	case multiraftv1.OperationResponseHeaders_CANCELED:
		return errors.NewCanceled(message)
	case multiraftv1.OperationResponseHeaders_NOT_FOUND:
		return errors.NewNotFound(message)
	case multiraftv1.OperationResponseHeaders_ALREADY_EXISTS:
		return errors.NewAlreadyExists(message)
	case multiraftv1.OperationResponseHeaders_UNAUTHORIZED:
		return errors.NewUnauthorized(message)
	case multiraftv1.OperationResponseHeaders_FORBIDDEN:
		return errors.NewForbidden(message)
	case multiraftv1.OperationResponseHeaders_CONFLICT:
		return errors.NewConflict(message)
	case multiraftv1.OperationResponseHeaders_INVALID:
		return errors.NewInvalid(message)
	case multiraftv1.OperationResponseHeaders_UNAVAILABLE:
		return errors.NewUnavailable(message)
	case multiraftv1.OperationResponseHeaders_NOT_SUPPORTED:
		return errors.NewNotSupported(message)
	case multiraftv1.OperationResponseHeaders_TIMEOUT:
		return errors.NewTimeout(message)
	case multiraftv1.OperationResponseHeaders_FAULT:
		return errors.NewFault(message)
	case multiraftv1.OperationResponseHeaders_INTERNAL:
		return errors.NewInternal(message)
	default:
		return errors.NewUnknown(message)
	}
}
