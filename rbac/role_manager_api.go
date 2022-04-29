// Copyright 2017 The casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rbac

import (
	"github.com/abichinger/govaluate"
)

type MatchingFunc func(arg1 string, arg2 string) bool

// RoleManager provides interface to define the operations for managing roles.
type IRoleManager interface {
	// Clear clears all stored data and resets the role manager to the initial state.
	Clear() error
	// AddLink adds the inheritance link between two roles. role: name1 and role: name2.
	// domain is a prefix to the roles (can be used for other purposes).
	AddLink(name1 string, name2 string, domain ...string) (bool, error)
	// DeleteLink deletes the inheritance link between two roles. role: name1 and role: name2.
	// domain is a prefix to the roles (can be used for other purposes).
	DeleteLink(name1 string, name2 string, domain ...string) (bool, error)
	// HasLink determines whether a link exists between two roles. role: name1 inherits role: name2.
	// domain is a prefix to the roles (can be used for other purposes).
	HasLink(name1 string, name2 string, domain ...string) (bool, error)
	// GetRoles gets the roles that a user inherits.
	// domain is a prefix to the roles (can be used for other purposes).
	GetRoles(name string, domain ...string) ([]string, error)
	// GetUsers gets the users that inherits a role.
	// domain is a prefix to the users (can be used for other purposes).
	GetUsers(name string, domain ...string) ([]string, error)

	SetMatcher(fn MatchingFunc)
	SetDomainMatcher(fn MatchingFunc)

	Range(fn func(name1, name2 string, domain ...string) bool)
}

// GenerateGFunction is the factory method of the g(_, _) function.
func GenerateGFunction(rm IRoleManager) govaluate.ExpressionFunction {

	return func(args ...interface{}) (interface{}, error) {
		name1 := args[0].(string)
		name2 := args[1].(string)

		if rm == nil {
			return name1 == name2, nil
		} else if len(args) == 2 {
			return rm.HasLink(name1, name2)
		} else {
			domains := []string{}
			for _, domain := range args[2:] {
				domains = append(domains, domain.(string))
			}
			return rm.HasLink(name1, name2, domains...)
		}
	}
}
