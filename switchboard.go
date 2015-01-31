package main

import (
	"sync"
)

type Post struct {
	Id   int
	Text string
}

type ChannelIndex string

type Subscription chan []byte
type Channel map[Subscription]bool
type ChannelMap struct {
	sync.RWMutex
	m map[ChannelIndex]Channel
}

type Switchboard struct {
	channels ChannelMap
}

func NewSwitchboard() *Switchboard {
	return &Switchboard{
		ChannelMap{m: make(map[ChannelIndex]Channel)},
	}
}

func (s *Switchboard) Unsubscribe(channel ChannelIndex, subscription Subscription) {
	s.channels.Lock()
	defer s.channels.Unlock()

	_, ok := s.channels.m[channel]
	if ok {
		_, ok := s.channels.m[channel][subscription]
		if ok {
			delete(s.channels.m[channel], subscription)
		}
	}
}

func (s *Switchboard) Subscribe(channel ChannelIndex) chan []byte {
	s.channels.Lock()
	defer s.channels.Unlock()

	subscription := make(Subscription)
	_, ok := s.channels.m[channel]
	if !ok {
		s.channels.m[channel] = make(Channel)
	}

	s.channels.m[channel][subscription] = true

	return subscription
}

func (s *Switchboard) Publish(channel ChannelIndex, post []byte) {
	s.channels.Lock()
	defer s.channels.Unlock()

	for subscription := range s.channels.m[channel] {
		subscription := subscription
		go func() {
			subscription <- []byte(post)
		}()
	}
}
