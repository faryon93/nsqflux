package main

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
	"errors"
	"strings"

	"github.com/faryon93/util"
	"github.com/nsqio/go-nsq"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/faryon93/nsqflux/payload"
	"github.com/faryon93/nsqflux/queue"
)

const (
	uriFromTopic   = 1
	uriFromChannel = 2
)

// ---------------------------------------------------------------------------------------
//  types
// ---------------------------------------------------------------------------------------

type Forwarder struct {
	Topic  string  `yaml:"topic" json:"topic"`
	Print  bool    `yaml:"print" json:"print"`
	Influx *Influx `yaml:"influx" json:"influx"`
}

// ---------------------------------------------------------------------------------------
//  public functions
// ---------------------------------------------------------------------------------------

func GetForwarders() ([]*Forwarder, error) {
	var forwarders []*Forwarder
	return forwarders, viper.UnmarshalKey("forward", &forwarders)
}

// ---------------------------------------------------------------------------------------
//  public members
// ---------------------------------------------------------------------------------------

func (f *Forwarder) IsValid() error {
	if strings.Count(f.Topic, ":") < 1 {
		return errors.New("from is invalidly formated")
	}

	if f.Influx == nil {
		return errors.New("influx configuration missing")
	}

	return nil
}

func (f *Forwarder) GetTopic() string {
	uri := strings.Split(f.Topic, ":")
	return uri[uriFromTopic]
}

func (f *Forwarder) GetChannel() string {
	uri := strings.Split(f.Topic, ":")
	if len(uri) < 3 {
		return util.RandomString(16) + "#ephemeral"
	}

	return uri[uriFromChannel]
}

func (f *Forwarder) Handle() queue.HandlerFunc {
	influxClient, err := f.Influx.GetClient()
	if err != nil {
		logrus.Errorln("failed to construct influx client:", err.Error())
		return nil
	}

	log := logrus.WithField("topic", f.GetTopic())
	parser := payload.NewParser(log, f.Influx.Database, f.Influx.Measurement)
	return func(msg *nsq.Message) error {
		batch, err := parser.ToBatchPoints(msg)
		if err != nil {
			// Return nil in order to discard the message.
			// This behaviour is ok, because whe parsing fails once
			// it will fail the next time too.
			log.Errorln("discarding invalid message:", err.Error())
			return nil
		}

		// Ignore empty batches. InfluxDB will throw an error otherwise.
		if batch == nil || len(batch.Points()) < 1 {
			return nil
		}

		// write to client
		err = influxClient.Write(batch)
		if err != nil {
			log.Errorln("influx write failed:", err.Error())
			return err
		}

		return nil
	}
}
