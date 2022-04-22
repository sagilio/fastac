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

package fastac

import (
	"errors"
	"fmt"

	"github.com/abichinger/fastac/model"
	m "github.com/abichinger/fastac/model"
	"github.com/abichinger/fastac/model/defs"
	"github.com/abichinger/fastac/model/effector"
	"github.com/abichinger/fastac/model/eft"
	"github.com/abichinger/fastac/model/matcher"
	"github.com/abichinger/fastac/model/types"
	"github.com/abichinger/fastac/storage"
	a "github.com/abichinger/fastac/storage/adapter"
	"github.com/abichinger/fastac/str"
)

type Enforcer struct {
	model   *m.Model
	adapter a.Adapter
	sc      *storage.StorageController
}

type Option func(*Enforcer) error

func OptionAutosave(autosave bool) Option {
	return func(e *Enforcer) error {
		if autosave {
			e.sc.EnableAutosave()
		} else {
			e.sc.DisableAutosave()
		}
		return nil
	}
}

func OptionStorage(enable bool) Option {
	return func(e *Enforcer) error {
		if enable {
			e.sc.Enable()
		} else {
			e.sc.Disable()
		}
		return nil
	}
}

func NewEnforcer(model interface{}, adapter interface{}, options ...Option) (*Enforcer, error) {
	e := &Enforcer{}

	switch model.(type) {
	case string:
		if m, err := m.NewModelFromFile(model.(string)); err != nil {
			return nil, err
		} else {
			e.model = m
		}
		break
	case m.Model:
		m2 := model.(m.Model)
		e.model = &m2
		break
	case *m.Model:
		e.model = model.(*m.Model)
		break
	default:
		return nil, errors.New(str.ERR_INVALID_MODEL)
	}

	var a2 a.Adapter
	switch adapter.(type) {
	case string:
		a2 := a.NewFileAdapter(adapter.(string))
		if err := a2.LoadPolicy(e.model); err != nil {
			return nil, err
		}
		break
	case a.Adapter:
		a2 = adapter.(a.Adapter)
		break
	default:
		a2 = &a.NoopAdapter{}
		options = append(options, OptionStorage(false))
	}

	e.SetAdapter(a2)

	for _, option := range options {
		if err := option(e); err != nil {
			return nil, err
		}
	}

	return e, nil
}

func (e *Enforcer) SetOption(option Option) error {
	return option(e)
}

func (e *Enforcer) SetAdapter(adapter a.Adapter) {
	autosave := false
	if e.sc != nil {
		autosave = e.sc.AutosaveEnabled()
		e.sc.Disable()
	}
	e.sc = storage.NewStorageController(e.model, adapter, autosave)
	e.adapter = adapter
}

func (e *Enforcer) LoadPolicy() error {
	if e.sc.Enabled() {
		e.sc.Disable()
		defer e.sc.Enable()
	}
	return e.adapter.LoadPolicy(e.model)
}

func (e *Enforcer) SavePolicy() error {
	return e.adapter.SavePolicy(e.model)
}

func (e *Enforcer) Flush() {
	e.sc.Flush()
}

func (e *Enforcer) AddRule(rule []string) (bool, error) {
	return e.model.AddRule(rule)
}

func (e *Enforcer) RemoveRule(rule []string) (bool, error) {
	return e.model.RemoveRule(rule)
}

func (e *Enforcer) AddRules(rules [][]string) error {
	if e.sc.AutosaveEnabled() {
		e.sc.DisableAutosave()
		defer e.sc.EnableAutosave()
		defer e.sc.Flush()
	}
	for _, rule := range rules {
		if _, err := e.model.AddRule(rule); err != nil {
			return err
		}
	}
	return nil
}

func (e *Enforcer) RemoveRules(rules [][]string) error {
	if e.sc.AutosaveEnabled() {
		e.sc.DisableAutosave()
		defer e.sc.EnableAutosave()
		defer e.sc.Flush()
	}
	for _, rule := range rules {
		if _, err := e.model.RemoveRule(rule); err != nil {
			return err
		}
	}
	return nil
}

func (e *Enforcer) Enforce(rvals ...interface{}) (bool, error) {
	return e.EnforceWithKeys("m", "r", "e", rvals...)
}

func (e *Enforcer) EnforceWithMatcher(matcher string, rvals ...interface{}) (bool, error) {
	return e.EnforceWithMatcherAndKeys(matcher, "r", "e", rvals...)
}

func (e *Enforcer) EnforceWithMatcherAndKeys(matcher string, rKey string, eKey string, rvals ...interface{}) (bool, error) {
	mKey := "m9999"
	e.model.AddDef('m', mKey, matcher)
	e.model.BuildMatcher(mKey)
	defer e.model.RemoveDef('m', mKey)
	return e.EnforceWithKeys(mKey, rKey, eKey, rvals...)
}

func (e *Enforcer) EnforceWithKeys(mKey string, rKey string, eKey string, rvals ...interface{}) (bool, error) {
	matcher, mOk := e.model.GetMatcher(mKey)
	if !mOk {
		return false, fmt.Errorf(str.ERR_MATCHER_NOT_FOUND, mKey)
	}
	rDef, rOk := e.model.GetRequestDef(rKey)
	if !rOk {
		return false, fmt.Errorf(str.ERR_REQUESTDEF_NOT_FOUND, rKey)
	}
	effector, eOk := e.model.GetEffector(eKey)
	if !eOk {
		return false, fmt.Errorf(str.ERR_EFFECTOR_NOT_FOUND, eKey)
	}

	return e.enforce(matcher, rDef, effector, rvals)
}

func (e *Enforcer) Filter(rvals ...interface{}) ([]types.Rule, error) {
	return e.FilterWithKeys("m", "r", rvals...)
}

func (e *Enforcer) FilterWithMatcher(matcher string, rvals ...interface{}) ([]types.Rule, error) {
	return e.FilterWithMatcherAndKeys(matcher, "r", rvals...)
}

func (e *Enforcer) FilterWithMatcherAndKeys(matcher string, rKey string, rvals ...interface{}) ([]types.Rule, error) {
	mKey := "m9999"
	e.model.AddDef('m', mKey, matcher)
	e.model.BuildMatcher(mKey)
	defer e.model.RemoveDef('m', mKey)
	return e.FilterWithKeys(mKey, rKey, rvals...)
}

func (e *Enforcer) FilterWithKeys(mKey string, rKey string, rvals ...interface{}) ([]types.Rule, error) {
	rules := []types.Rule{}
	err := e.RangeMatchesWithKeys(mKey, rKey, rvals, func(rule types.Rule) bool {
		rules = append(rules, rule)
		return true
	})
	return rules, err
}

func (e *Enforcer) RangeMatches(matcher *matcher.Matcher, rDef *defs.RequestDef, rvals []interface{}, fn func(rule types.Rule) bool) error {
	return e.model.RangeMatches(matcher, rDef, rvals, fn)
}

func (e *Enforcer) RangeMatchesWithKeys(mKey string, rKey string, rvals []interface{}, fn func(rule types.Rule) bool) error {
	matcher, mOk := e.model.GetMatcher(mKey)
	if !mOk {
		return fmt.Errorf(str.ERR_MATCHER_NOT_FOUND, mKey)
	}
	rDef, rOk := e.model.GetRequestDef(rKey)
	if !rOk {
		return fmt.Errorf(str.ERR_REQUESTDEF_NOT_FOUND, rKey)
	}

	return e.RangeMatches(matcher, rDef, rvals, fn)
}

func (e *Enforcer) enforce(matcher *matcher.Matcher, rDef *defs.RequestDef, effector effector.Effector, rvals []interface{}) (bool, error) {
	pDef := matcher.GetPolicy()
	res := eft.Indeterminate
	effects := []types.Effect{}
	matches := []types.Rule{}

	var eftErr error = nil
	err := e.RangeMatches(matcher, rDef, rvals, func(rule types.Rule) bool {
		effect := pDef.GetEft(rule)

		effects = append(effects, effect)
		matches = append(matches, rule)

		res, _, eftErr = effector.MergeEffects(effects, matches, false)

		if eftErr != nil || res != eft.Indeterminate {
			return false
		}
		return true
	})
	if err != nil {
		return false, err
	}
	if eftErr != nil {
		return false, err
	}

	if res == eft.Indeterminate {
		res, _, _ = effector.MergeEffects(effects, matches, true)
	}

	return res == eft.Allow, nil
}

func (e *Enforcer) GetModel() *model.Model {
	return e.model
}
