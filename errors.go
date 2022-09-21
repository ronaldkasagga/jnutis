package jnutis

import (
	"github.com/pkg/errors"
)

type IdMap map[int64]struct{}

func (m IdMap) list() (out []int64) {
	if m.empty() {
		return
	}
	var idx int
	out = make([]int64, m.size())
	for id := range m {
		out[idx] = id
		idx++
	}
	return
}

func (m IdMap) size() int {
	return len(m)
}

func (m IdMap) empty() bool {
	return m.size() == 0
}

type ErrorMap map[string]IdMap

func (l *ErrorMap) add(errString string, ids ...int64) *ErrorMap {
	if len(ids) == 0 || errString == "" {
		return l
	} else if _, ok := (*l)[errString]; !ok {
		(*l)[errString] = make(IdMap, 0)
	}
	for _, id := range ids {
		(*l)[errString][id] = struct{}{}
	}
	return l
}

func (l *ErrorMap) merge(list ErrorMap) *ErrorMap {
	if len(list) == 0 {
		return l
	}
	for er, ids := range list {
		l.add(er, ids.list()...)
	}
	return l
}

func (l *ErrorMap) Error(subject string) error {
	if !l.HasErrors() {
		return nil
	}
	var total int
	for e := range *l {
		total += (*l)[e].size()
	}
	return errors.Errorf("encountered %d errors with %d %s", l.Size(), total, subject)
}

func (l *ErrorMap) HasErrors() bool {
	return l.Size() > 0
}

func (l *ErrorMap) None() bool {
	return !l.HasErrors()
}

func (l *ErrorMap) Size() int {
	return len(*l)
}

func (l *ErrorMap) Data() (out map[string][]int64) {
	if l.None() {
		return
	}
	out = make(map[string][]int64, l.Size())
	for e, m := range *l {
		out[e] = m.list()
	}
	return
}
