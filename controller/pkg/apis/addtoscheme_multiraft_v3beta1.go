// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package apis

import (
	storagev2beta2 "github.com/atomix/consensus/controller/pkg/apis/multiraft/v3beta1"
)

func init() {
	// register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, storagev2beta2.SchemeBuilder.AddToScheme)
}
