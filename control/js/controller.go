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
	code, err := ctrl.getCode()
	if err != nil {
		return nil, err
	}
	Otto := otto.New()
	ctrl.registerBuiltins(Otto)

	value, err := Otto.Run(code)
	if err != nil {
		return nil, err
	}
	return ctrl.resultToInstructions(&value)
}

func (ctrl *JsController) getCode() (string, error) {
	return `result = {upstreams: [{url: "http://localhost:5000"}]}`, nil
}

func (ctrl *JsController) resultToInstructions(value *otto.Value) (*ProxyInstructions, error) {
	obj, err := value.Export()
	if err != nil {
		return nil, err
	}
	glog.Infof("Got value %#v -exported into-> %#v", value, obj)
	response, ok := obj.(map[string]interface{})
	if !ok {
		glog.Errorf("Invalid response from json server, expected dictionary, got %#v", obj)
		return nil, fmt.Errorf("Internal error")
	}

	return ProxyInstructionsFromObject(response)
}

func (ctrl *JsController) registerBuiltins(o *otto.Otto) {
	ctrl.addDiscoveryService(o)
}

func (ctrl *JsController) addDiscoveryService(o *otto.Otto) {
	o.Set("discover", func(call otto.FunctionCall) otto.Value {
		right, _ := call.Argument(0).ToString()
		value, err := ctrl.DiscoveryService.Get(right)
		glog.Infof("Got %v, %s", value, err)
		result, _ := o.ToValue(value)
		return result
	})
}
