// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package v1_26

import "xorm.io/xorm"

// AddParentIDToUser adds the parent_id column to the user table to support
// organization subgroups (nested organizations).
func AddParentIDToUser(x *xorm.Engine) error {
	type User struct {
		ParentID int64 `xorm:"NOT NULL DEFAULT 0"`
	}
	return x.Sync(new(User))
}
