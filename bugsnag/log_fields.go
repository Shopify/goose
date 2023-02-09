package bugsnag

import (
	bugsnaggo "github.com/bugsnag/bugsnag-go/v2"
	log "github.com/sirupsen/logrus"
)

func LogFieldsHook(event *bugsnaggo.Event, config *bugsnaggo.Configuration) error {
	for _, datum := range event.RawData {
		switch d := datum.(type) {
		case *log.Entry:
			if event.Ctx == nil {
				event.Ctx = d.Context
			}
			if event.Context == "" && event.Message != d.Message {
				event.Context = d.Message
			}
			event.MetaData.Update(bugsnaggo.MetaData{
				"log": {
					"message": d.Message,
				},
				"metadata": d.Data,
			})
		case log.Fields:
			event.MetaData.Update(bugsnaggo.MetaData{
				"metadata": d,
			})
		}
	}

	return nil
}
