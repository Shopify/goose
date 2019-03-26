package bugsnag

import (
	bugsnaggo "github.com/bugsnag/bugsnag-go"
)

// bugsnagdata is the utility object which helps with writing bugsnag's unit tests
type bugsnagdata struct {
	dataList []interface{}
}

func newBugsnagData(rawData []interface{}) *bugsnagdata {
	return &bugsnagdata{
		dataList: rawData,
	}
}

func (dl *bugsnagdata) getBugsnaggoErrorClass() string {
	for _, data := range dl.dataList {
		if errObj, ok := data.(bugsnaggo.ErrorClass); ok {
			return errObj.Name
		}
	}
	return ""
}

func (dl *bugsnagdata) hasErrTab() bool {
	for _, data := range dl.dataList {
		if myMap, ok := data.(bugsnaggo.MetaData); ok {
			if v, ok := myMap["Error"]; ok {
				if _, ok := v["details"]; ok {
					return true
				}
			}
		}
	}
	return false
}

func (dl *bugsnagdata) getLogTab() Rows {
	for _, data := range dl.dataList {
		if myMap, ok := data.(bugsnaggo.MetaData); ok {
			if t, ok := myMap["Log"]; ok {
				return t
			}
		}
	}
	return nil
}

func (dl *bugsnagdata) getContext() string {
	for _, data := range dl.dataList {
		if cont, ok := data.(bugsnaggo.Context); ok {
			return cont.String
		}
	}
	return ""
}

func (dl *bugsnagdata) getUser() string {
	for _, data := range dl.dataList {
		if user, ok := data.(bugsnaggo.User); ok {
			return user.Id
		}
	}
	return ""
}

func (dl *bugsnagdata) getTab(label string) Rows {
	for _, data := range dl.dataList {
		if myMap, ok := data.(bugsnaggo.MetaData); ok {
			if t, ok := myMap[label]; ok {
				return t
			}
		}
	}
	return nil
}

func (dl *bugsnagdata) hasValue(data interface{}) bool {
	for _, v := range dl.dataList {
		if v == data {
			return true
		}
	}
	return false
}
