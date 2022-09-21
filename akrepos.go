package jnutis

import (
	"github.com/sirupsen/logrus"
"github.com/gertd/go-pluralize"
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

func ProcessEvents[T Model](log *logrus.Entry, subject string, processFn ProcessFunc[T], data []T) (failed ErrorMap) {
	client := pluralize.NewClient()
	failed = make(ErrorMap)
	if size := len(data); size == 0 {
		log.Warnf("received no %s. doing nothing", client.Plural(subject))
		return
	} else if size == 1 {
		log.WithField(client.Singular(subject), data[0].Identifier()).Info("retrying final event in branch")
		if e := processFn(data); e != nil {
			log.WithField(client.Singular(subject), data[0].Identifier()).
				WithError(e).Info("operation failed")
			failed.add(e.Error(), data[0].Identifier())
		}
		return
	}

	mid := int(len(data) / 2)
	splitData := [][]T{data[:mid], data[mid:]}
	log.WithField(client.Plural(subject), map[string][]int64{
		"left":  getIdentifiers(splitData[0]),
		"right": getIdentifiers(splitData[1]),
	}).Infof("spliting %s for retry", client.Plural(subject))
	for i := range splitData {
		if e := processFn(splitData[i]); e == nil {
			continue
		} else if len(splitData[i]) == 1 {
			log.WithField(client.Singular(subject), splitData[i][0].Identifier()).
				WithError(e).Info("operation failed")
			failed.add(e.Error(), splitData[i][0].Identifier())
		} else {
			log.WithField(client.Plural(subject), getIdentifiers(splitData[i])).
				WithError(e).Info("retrying %s recursively", client.Plural(subject))
			failed.merge(ProcessEvents(log, subject, processFn, splitData[i]))
		}
	}
	return
}
