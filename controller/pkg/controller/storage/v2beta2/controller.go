// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package v2beta2

import (
	"github.com/atomix/runtime/sdk/pkg/logging"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var log = logging.GetLogger()

// AddControllers creates a new Partition ManagementGroup and adds it to the Manager. The Manager will set fields on the ManagementGroup
// and Start it when the Manager is Started.
func AddControllers(mgr manager.Manager) error {
	if err := addMultiRaftStoreController(mgr); err != nil {
		return err
	}
	if err := addMultiRaftClusterController(mgr); err != nil {
		return err
	}
	if err := addMultiRaftNodeController(mgr); err != nil {
		return err
	}
	if err := addRaftGroupController(mgr); err != nil {
		return err
	}
	if err := addRaftMemberController(mgr); err != nil {
		return err
	}
	return nil
}
