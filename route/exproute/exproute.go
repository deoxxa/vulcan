/*
Expression based request router, supports functions and combinations of functions in form

<What to match><Matching verb> and || and && operators.

*/
package exproute

import (
	"github.com/mailgun/vulcan/location"
	"github.com/mailgun/vulcan/request"
	"sync"
)

type ExpRouter struct {
	mutex *sync.Mutex
	tree  *matchTree
}

func NewExpRouter() *ExpRouter {
	return &ExpRouter{
		mutex: &sync.Mutex{},
	}
}

func (e *ExpRouter) Route(req request.Request) (location.Location, error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	return nil, nil
}
