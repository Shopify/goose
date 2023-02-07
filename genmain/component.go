package genmain

import "github.com/Shopify/goose/v2/safely"

// Component is used to represent various "components". At a high level, main()
// essentially cobbles together a few components whose lifecycles are managed
// by Tombs. `Component` allows us to treat them as black boxes.
type Component interface {
	safely.Runnable
}

type ComponentWithDependencies interface {
	Component

	Dependencies() []Component
}

func NewComponentWithDependencies(c Component, dependencies ...Component) ComponentWithDependencies {
	return &dependencyWrapper{c, dependencies}
}

type dependencyWrapper struct {
	Component
	dependencies []Component
}

func (w *dependencyWrapper) Dependencies() []Component {
	return w.dependencies
}
