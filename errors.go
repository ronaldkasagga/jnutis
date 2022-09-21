package jnutis

import (
	"fmt"
	"strings"

	"github.com/gertd/go-pluralize"
	"github.com/pkg/errors"
)

type IdMap map[int64]struct{}

type Subject interface {
}

func subjectToString(s Subject) (sub string) {
	sub = fmt.Sprintf("%T", s)
	if strings.Contains(sub, ".") {
		return
	} else if idx := strings.Index(sub, "."); idx != -1 {
		return sub[idx:]
	}
	return
}

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

type ErrorMap[subject Subject] map[string]IdMap

func (l *ErrorMap[subject]) add(errString string, ids ...int64) *ErrorMap[subject] {
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

func (l *ErrorMap[subject]) merge(list ErrorMap[subject]) *ErrorMap[subject] {
	if len(list) == 0 {
		return l
	}
	for er, ids := range list {
		l.add(er, ids.list()...)
	}
	return l
}

func (l *ErrorMap[subject]) Error() error {
	if !l.HasErrors() {
		return nil
	}
	var total int
	for e := range *l {
		total += (*l)[e].size()
	}
	client := pluralize.NewClient()
	var s subject
	var model string = subjectToString(s)
	size := l.Size()
	return errors.Errorf("encountered %s with %s", 
		client.Pluralize("error", size, true),
		client.Pluralize(model, size, true),
	)
}

func (l *ErrorMap[subject]) HasErrors() bool {
	return l.Size() > 0
}

func (l *ErrorMap[subject]) None() bool {
	return !l.HasErrors()
}

func (l *ErrorMap[subject]) Size() int {
	return len(*l)
}

func (l *ErrorMap[subject]) Data() (out map[string][]int64) {
	if l.None() {
		return
	}
	out = make(map[string][]int64, l.Size())
	for e, m := range *l {
		out[e] = m.list()
	}
	return
}
