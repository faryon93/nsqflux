package queue

// nsqflux
// Copyright (C) 2019 Maximilian Pachl

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

// ---------------------------------------------------------------------------------------
//  imports
// ---------------------------------------------------------------------------------------

import (
	"sync"
	"time"

	"github.com/nsqio/go-nsq"
)

// ---------------------------------------------------------------------------------------
//  constants
// ---------------------------------------------------------------------------------------

const (
	PollInterval = 20 * time.Second
	RequeueDelay = 10 * time.Second
	MaxInFlight  = 32
)

var (
	ClientId = ""
)

// ---------------------------------------------------------------------------------------
// types
// ---------------------------------------------------------------------------------------

type Queue struct {
	LookupdAddr  []string
	PollInterval time.Duration
	MaxInFlight  int

	// private fields
	m         sync.Mutex
	consumers []*nsq.Consumer
}

type HandlerFunc func(message *nsq.Message) error

// ---------------------------------------------------------------------------------------
//  public members
// ---------------------------------------------------------------------------------------

func New(lookupd ...string) *Queue {
	return &Queue{
		LookupdAddr:  lookupd,
		PollInterval: PollInterval,
		MaxInFlight:  MaxInFlight,

		consumers: make([]*nsq.Consumer, 0),
	}
}

// ---------------------------------------------------------------------------------------
//  public members
// ---------------------------------------------------------------------------------------

func (q *Queue) Subscribe(topic, channel string, fn HandlerFunc) error {
	cfg := nsq.NewConfig()
	cfg.ClientID = ClientId
	cfg.LookupdPollInterval = q.PollInterval
	cfg.MaxInFlight = q.MaxInFlight
	cfg.DefaultRequeueDelay = RequeueDelay

	consumer, err := nsq.NewConsumer(topic, channel, cfg)
	if err != nil {
		return err
	}
	consumer.AddConcurrentHandlers(nsq.HandlerFunc(fn), cfg.MaxInFlight)
	consumer.SetLogger(&logger{}, nsq.LogLevelError)

	err = consumer.ConnectToNSQLookupds(q.LookupdAddr)
	if err != nil {
		return err
	}

	q.m.Lock()
	q.consumers = append(q.consumers)
	q.m.Unlock()

	return nil
}

func (q *Queue) Stop() {
	q.m.Lock()
	defer q.m.Unlock()

	for _, consumer := range q.consumers {
		consumer.Stop()
	}
}
