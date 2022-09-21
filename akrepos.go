package jnutis

import (
	"github.com/gertd/go-pluralize"
	"github.com/sirupsen/logrus"
)

type Model interface {
	Identifier() int64
}

type ProcessFunc[T Model] func([]T) error

func getIdentifiers[T Model](data []T) (out []int64) {
	if len(data) == 0 {
		return
	}
	out = make([]int64, len(data))
	for i, d := range data {
		out[i] = d.Identifier()
	}
	return
}

func ProcessWithSplitRetry[subject Subject, T Model](log *logrus.Entry, processFn ProcessFunc[T], data []T) (failed ErrorMap[subject]) {
	client := pluralize.NewClient()
	failed = make(ErrorMap[subject])
	var s subject
	var model string = subjectToString(s)
	if size := len(data); size == 0 {
		log.Warnf("received no %s. doing nothing", client.Plural(model))
		return
	} else if size == 1 {
		log.WithField(client.Singular(model), data[0].Identifier()).Info("retrying final event in branch")
		if e := processFn(data); e != nil {
			log.WithField(client.Singular(model), data[0].Identifier()).
				WithError(e).Info("operation failed")
			failed.add(e.Error(), data[0].Identifier())
		}
		return
	}

	mid := int(len(data) / 2)
	splitData := [][]T{data[:mid], data[mid:]}
	log.WithField(client.Plural(model), map[string][]int64{
		"left":  getIdentifiers(splitData[0]),
		"right": getIdentifiers(splitData[1]),
	}).Infof("spliting %s for retry", client.Plural(model))
	for i := range splitData {
		if e := processFn(splitData[i]); e == nil {
			continue
		} else if len(splitData[i]) == 1 {
			log.WithField(client.Singular(model), splitData[i][0].Identifier()).
				WithError(e).Info("operation failed")
			failed.add(e.Error(), splitData[i][0].Identifier())
		} else {
			log.WithField(client.Plural(model), getIdentifiers(splitData[i])).
				WithError(e).Info("retrying %s recursively", client.Plural(model))
			failed.merge(ProcessWithSplitRetry[subject](log, processFn, splitData[i]))
		}
	}
	return
}
