package main

import "testing"

func TestNewSwitchboard(t *testing.T) {
	s := NewSwitchboard()

	if s == nil {
		t.Error("Expected NewSwitchboard() to return a non-nil value")
	}
}

func TestSubscribe(t *testing.T) {
	s := NewSwitchboard()
	ch := s.Subscribe(1)

	var v []byte
	go func() {
		v = <-ch
	}()

	ch <- []byte("hello")

	if string(v) != "hello" {
		t.Errorf("Expected %v == \"hello\"")
	}
}

func TestUnsubscribe(t *testing.T) {
	s := NewSwitchboard()
	ch := s.Subscribe(1)

	s.Unsubscribe(1, ch)
}
