package jnutis

import (

	"github.com/gertd/go-pluralize"
	"github.com/pkg/errors"
)

func ModelErrorMap(model string) ErrorMap {
	return &_errorMap{data: make(errMap), model: model}
}

type ErrorMap interface {
	add(errString string, ids ...int64) ErrorMap
	merge(list ErrorMap) ErrorMap
	Error() error
	HasErrors() bool
	None() bool
	Size() int
	Data() (out map[string][]int64)
}

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

type errMap map[string]IdMap

func (l *errMap) add(errString string, ids ...int64) *errMap {
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

func (l *errMap) merge(list errMap) *errMap {
	if len(list) == 0 {
		return l
	}
	for er, ids := range list {
		l.add(er, ids.list()...)
	}
	return l
}

func (l *errMap) Error(model string) error {
	if !l.HasErrors() {
		return nil
	}
	var total int
	for e := range *l {
		total += (*l)[e].size()
	}
	client := pluralize.NewClient()
	size := l.Size()
	return errors.Errorf("encountered %s with %s",
		client.Pluralize("error", size, true),
		client.Pluralize(model, size, true),
	)
}

func (l *errMap) HasErrors() bool {
	return l.Size() > 0
}

func (l *errMap) None() bool {
	return !l.HasErrors()
}

func (l *errMap) Size() int {
	return len(*l)
}

func (l *errMap) Data() (out map[string][]int64) {
	if l.None() {
		return
	}
	out = make(map[string][]int64, l.Size())
	for e, m := range *l {
		out[e] = m.list()
	}
	return
}
type _errorMap struct {
	data errMap
	model string
}

func (l *_errorMap) add(errString string, ids ...int64) ErrorMap {
	if len(l.data) != 0 {
		l.data.add(errString, ids...)
	}
	return l
}

func (l *_errorMap) merge(list ErrorMap) ErrorMap {
	if len(l.data) > 0 && list != nil || list.HasErrors() {
		l.data.merge(list.(*_errorMap).data)
	}
	return l
}

func (l *_errorMap) Error() error {
	if len(l.data) > 0 {
		return l.data.Error(l.model)
	}
	return nil
}

func (l *_errorMap) HasErrors() bool {
	return len(l.data) > 0 && l.data.HasErrors()
}

func (l *_errorMap) None() bool {
	return len(l.data) == 0 || l.data.None()
}

func (l *_errorMap) Size() int {
	if len(l.data) == 0 {
		return 0
	}
	return l.data.Size()
}

func (l *_errorMap) Data() (out map[string][]int64) {
	if len(l.data) == 0 {
		return
	}
	return l.data.Data()
}
