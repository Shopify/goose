package bugsnag

import bugsnaggo "github.com/bugsnag/bugsnag-go/v2"

// TabProvider defines the interface to be implemented by the different client objects that want custom decorated bugsnag tabs
type TabProvider interface {
	CreateBugsnagTab() Tab
}

// Tab can be used to attach an additional info tab to bugsnag error reports.
// Just pass it to the Notify function.
type Tab struct {
	Label string
	Rows  Rows
}

// Rows is a helper type to make it easier to build bugsnag tab contents,
// with simple Go literals as in `Tab{"Stuff",Rows{"a":1,"b":2,...}}`.
type Rows map[string]interface{}

func TabHook(event *bugsnaggo.Event, config *bugsnaggo.Configuration) error {
	for _, datum := range event.RawData {
		switch d := datum.(type) {
		case TabProvider:
			tab := d.CreateBugsnagTab()
			event.MetaData.Update(bugsnaggo.MetaData{
				tab.Label: tab.Rows,
			})
		case Tab:
			event.MetaData.Update(bugsnaggo.MetaData{
				d.Label: d.Rows,
			})
		}
	}

	return nil
}
