package js

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/mailgun/vulcan/discovery"
	. "github.com/mailgun/vulcan/instructions"
	"github.com/robertkrimen/otto"
	"net/http"
)

type JsController struct {
	DiscoveryService discovery.Service
}

func (ctrl *JsController) GetInstructions(req *http.Request) (instr *ProxyInstructions, err error) {
	instr = nil
	err = fmt.Errorf("Not implemented")
	defer func() {
		if r := recover(); r != nil {
			glog.Errorf("Recovered:", r)
			err = fmt.Errorf("Internal js error")
			instr = nil
		}
	}()
	Otto := otto.New()
	Otto.Set("getKey", func(call otto.FunctionCall) otto.Value {
		right, _ := call.Argument(0).ToString()
		value, err := ctrl.DiscoveryService.Get(right)
		glog.Infof("Got %v, %s", value, err)
		result, _ := Otto.ToValue(value)
		return result
	})

	value, err := Otto.Run(`
         [getKey("upstream")]
     `)
	if err != nil {
		return nil, err
	}
	obj, err := value.Export()
	if err != nil {
		return nil, err
	}
	values := obj.([]interface{})
	upstream, _ := values[0].(string)

	u, err := NewUpstream(upstream, []*Rate{}, map[string][]string{})
	if err != nil {
		return nil, err
	}
	return &ProxyInstructions{
		Upstreams: []*Upstream{u},
	}, nil
}
