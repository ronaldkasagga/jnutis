package jnutis

import (
	"github.com/gertd/go-pluralize"
	"github.com/sirupsen/logrus"
)

type Model interface {
	Identifier() int64
}

type ProcessFunc func([]Model) error

func getIdentifiers(data []Model) (out []int64) {
	if len(data) == 0 {
		return
	}
	out = make([]int64, len(data))
	for i, d := range data {
		out[i] = d.Identifier()
	}
	return
}

func ProcessWithSplitRetry(log *logrus.Entry, model string, processFn ProcessFunc, data []Model) (failed ErrorMap) {
	client := pluralize.NewClient()
	failed = &_errorMap{data: make(errMap), model: model}
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
	splitData := [][]Model{data[:mid], data[mid:]}
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
			failed.merge(ProcessWithSplitRetry(log, model, processFn, splitData[i]))
		}
	}
	return
}
