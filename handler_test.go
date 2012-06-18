package congo

import (
	"reflect"
	"testing"
)

type AlternateContext struct {
	Context
}

type UnknownResponse struct {
}

var appliedOrder []string
var altContextAssertion bool = false
var nullResponse *NullResponse = &NullResponse{}

func initAppliedOrder() {
	appliedOrder = make([]string, 0)
}

func appendAppliedOrder(val string) {
	appliedOrder = append(appliedOrder, val)
}

func ActionA(c Context) (Context, interface{}) {
	appendAppliedOrder("A")
	return c, nil
}

func ActionB(c Context) (Context, interface{}) {
	appendAppliedOrder("B")
	return c, nil
}

func ActionC(c Context) (Context, interface{}) {
	appendAppliedOrder("C")
	return c, nullResponse
}

func ActionAltContext(c Context) (Context, interface{}) {
	alt := &AlternateContext{c}
	return alt, nil
}

func ActionPostAltContext(c Context) (Context, interface{}) {
	_, altContextAssertion = c.(*AlternateContext)
	return c, nullResponse
}

func ActionUnknownResponse(c Context) (Context, interface{}) {
	return c, &UnknownResponse{}
}

func TestActionsAppliedInOrder(t *testing.T) {
	initAppliedOrder()

	handler := NewHandler()
	handler.Actions(ActionB, ActionA, ActionC)
	handlerFn := MuxHandler(handler)
	handlerFn(nil, nil)

	expected := []string{"B", "A", "C"}
	if !reflect.DeepEqual(appliedOrder, expected) {
		t.Fail()
	}
}

func TestCopyRetainsActionOrder(t *testing.T) {
	initAppliedOrder()

	source := NewHandler()
	source.Actions(ActionB, ActionC)
	sourceFn := MuxHandler(source)
	sourceFn(nil, nil)
	sourceOrder := make([]string, len(appliedOrder))
	copy(sourceOrder, appliedOrder)

	initAppliedOrder()

	copy := source.Copy()
	copyFn := MuxHandler(copy)
	copyFn(nil, nil)

	if !reflect.DeepEqual(appliedOrder, sourceOrder) {
		t.Fail()
	}
}

func TestAlternateContextPassesThrough(t *testing.T) {
	handler := NewHandler()
	handler.Actions(ActionAltContext, ActionPostAltContext)
	handlerFn := MuxHandler(handler)
	handlerFn(nil, nil)

	if !altContextAssertion {
		t.Fail()
	}

	altContextAssertion = false
}

func TestNoResponseReturnedFromChainPanics(t *testing.T) {
	handler := NewHandler()
	handler.Actions(ActionA)
	handlerFn := MuxHandler(handler)

	defer func() {
		if e := recover(); e == nil {
			t.Fail()
		}
	}()

	handlerFn(nil, nil)
}

func TestUnknownResponseCausesPanic(t *testing.T) {
	handler := NewHandler()
	handler.Actions(ActionUnknownResponse)
	handlerFn := MuxHandler(handler)

	defer func() {
		if e := recover(); e == nil {
			t.Fail()
		}
	}()

	handlerFn(nil, nil)
}
