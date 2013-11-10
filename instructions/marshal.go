/*
This module contains logic unmarshalling rates, upstreams and tokens
from json encoded strings. This logic is pretty verbose, so we are
concentrating it here to keep original modules clean and focusd on the
acutal actions.
*/
package instructions

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type ProxyInstructionsObj struct {
	Failover  Failover
	Tokens    []TokenObj
	Upstreams []UpstreamObj
	Headers   map[string][]string
}

type TokenObj struct {
	Id    string
	Rates []RateObj
}

type RateObj struct {
	Increment int64
	Value     int64
	Period    string
}

// This object is used for unmarshalling from json
type UpstreamObj struct {
	Url     string
	Rates   []RateObj
	Headers map[string][]string
}

func ProxyInstructionsFromJson(bytes []byte) (*ProxyInstructions, error) {
	var r interface{}
	err := json.Unmarshal(bytes, &r)
	if err != nil {
		return nil, err
	}
	return ProxyInstructionsFromObj(r)
}

func ProxyInstructionsFromObj(in interface{}) (*ProxyInstructions, error) {
	obj, ok := in.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Expected dictionary as instructions")
	}

	// Parse upstreams
	upstreamsObj, exists := obj["upstreams"]
	if !exists {
		return nil, fmt.Errorf("Expected upstreams")
	}
	upstreamsList, ok := upstreamsObj.([]interface{})
	if !ok {
		return nil, fmt.Errorf("Upstreams should be a list")
	}
	upstreams, err := upstreamsFromList(upstreamsList)
	if err != nil {
		return nil, err
	}

	// Rate-limiting tokens
	tokensObj, exists := obj["tokens"]
	if !exists {
		return nil, fmt.Errorf("Expected tokens")
	}
	tokensList, ok := tokensObj.([]interface{})
	if !ok {
		return nil, fmt.Errorf("Tokens should be a list")
	}
	tokens, err := tokensFromList(tokensList)
	if err != nil {
		return nil, err
	}

	// Failover instructions
	var failover *Failover
	failoverObj, exists := obj["failover"]
	if exists {
		failover, err = failoverFromObj(failoverObj)
		if err != nil {
			return nil, err
		}
	}

	// Request headers
	var headers http.Header
	headersObj, exists := obj["headers"]
	if exists {
		failover, err = headersFromObj(headersObj)
		if err != nil {
			return nil, err
		}
	}

	return NewProxyInstructions(failover, tokens, upstreams, headers)
}

func upstreamsFromList(inUpstreams []interface{}) ([]*Upstream, error) {
	upstreams := make([]*Upstream, len(inUpstreams))
	for i, upstreamObj := range inUpstreams {
		upstream, err := upstreamFromObj(upstreamObj)
		if err != nil {
			return nil, err
		}
		upstreams[i] = upstream
	}
	return upstreams, nil
}

func upstreamFromObj(in interface{}) (*Upstream, error) {
	obj, ok := in.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Upstream should be dictionary")
	}
	url, ok := obj["url"]
	if !ok {
		return nil, fmt.Errorf("Expected url")
	}
}

func tokensFromList(inTokens []interface{}) ([]*Token, error) {
	tokens := make([]*Token, len(inTokens))
	for i, tokenObj := range inTokens {
		rates, err := tokenFromObj(tokenObj)
		if err != nil {
			return nil, err
		}
		tokens[i] = token
	}
	return tokens, nil
}

func rateFromObj(obj RateObj) (*Rate, error) {
	period, err := periodFromString(obj.Period)
	if err != nil {
		return nil, err
	}
	return NewRate(obj.Increment, obj.Value, period)
}

//helper to unmarshal periods to golang time.Duration
func periodFromString(period string) (time.Duration, error) {
	switch strings.ToLower(period) {
	case "second":
		return time.Second, nil
	case "minute":
		return time.Minute, nil
	case "hour":
		return time.Hour, nil
	}
	return -1, fmt.Errorf("Unsupported period: %s", period)
}
