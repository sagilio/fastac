// Copyright 2022 The FastAC Authors. All Rights Reserved.
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

package policy

import (
	"github.com/abichinger/fastac/model/defs"
	"github.com/abichinger/fastac/model/types"
	"github.com/abichinger/fastac/util"
	em "github.com/vansante/go-event-emitter"
)

type Policy struct {
	ruleMap map[string]types.Rule

	*em.Emitter
	*defs.PolicyDef
}

func NewPolicy(pDef *defs.PolicyDef) *Policy {
	p := &Policy{}
	p.PolicyDef = pDef
	p.Emitter = em.NewEmitter(false)
	p.ruleMap = make(map[string]types.Rule)
	return p
}

func (p *Policy) AddRule(rule []string) (bool, error) {
	key := util.Hash(rule)
	if _, ok := p.ruleMap[key]; ok {
		return false, nil
	}
	p.ruleMap[key] = rule
	p.Emitter.EmitEvent(EVT_RULE_ADDED, rule)
	return true, nil
}

func (p *Policy) RemoveRule(rule []string) (bool, error) {
	key := util.Hash(rule)
	_, ok := p.ruleMap[key]
	if !ok {
		return false, nil
	}
	delete(p.ruleMap, key)
	p.Emitter.EmitEvent(EVT_RULE_REMOVED, rule)
	return true, nil
}

func (p *Policy) Range(fn func(rule []string) bool) {
	for _, r := range p.ruleMap {
		if !fn(r) {
			break
		}
	}
}

func (p *Policy) GetDistinct(columns []int) ([][]string, error) {
	return GetDistinct(p, columns)
}

func (p *Policy) Clear() error {
	p.ruleMap = make(map[string]types.Rule)
	p.Emitter.EmitEvent(EVT_CLEARED)
	return nil
}
