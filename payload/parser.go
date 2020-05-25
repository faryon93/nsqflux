package payload

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
	"fmt"
	"strings"
	"time"

	influx "github.com/influxdata/influxdb1-client/v2"
	"github.com/nsqio/go-nsq"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fastjson"
)

const (
	KeyMeasurement = "_measurement"
	KeyTimestamp   = "_timestamp"

	FieldOverride = "$"
)

// ---------------------------------------------------------------------------------------
//  imports
// ---------------------------------------------------------------------------------------

var (
	parsers fastjson.ParserPool
)

// ---------------------------------------------------------------------------------------
//  public functions
// ---------------------------------------------------------------------------------------

type Parser struct {
	log         *logrus.Entry
	database    string
	measurement string
}

// ---------------------------------------------------------------------------------------
//  public functions
// ---------------------------------------------------------------------------------------

func NewParser(log *logrus.Entry, database string, measurement string) *Parser {
	return &Parser{
		log:         log,
		database:    database,
		measurement: measurement,
	}
}

func (p *Parser) ToBatchPoints(msg *nsq.Message) (influx.BatchPoints, error) {
	json, err := parsers.Get().ParseBytes(msg.Body)
	if err != nil {
		return nil, err
	}

	// process the request
	batch, _ := influx.NewBatchPoints(influx.BatchPointsConfig{
		Precision: "ms",
		Database:  p.database,
	})
	p.process(batch, json, time.Unix(0, msg.Timestamp))

	return batch, nil
}

// ---------------------------------------------------------------------------------------
//  private members
// ---------------------------------------------------------------------------------------

// process addes all datapoints, which are present in v to batch.
func (p *Parser) process(bt influx.BatchPoints, v *fastjson.Value, time time.Time) {
	measurement := p.measurement

	switch v.Type() {
	case fastjson.TypeObject:
		obj, _ := v.Object()

		// override the default measurment name, if the special field exists
		if m := obj.Get(KeyMeasurement); m != nil && m.Type() == fastjson.TypeString {
			measurement = m.String()
		}

		// override the timestamp, if special field exits
		timestamp, wrn := getTimestamp(obj, time)
		if wrn != nil {
			p.log.Errorln(wrn.Error())
		}

		// extract all strings of the json object as tags
		// and all numbers as field values (as float)
		tags := make(map[string]string)
		fields := make(map[string]interface{})
		obj.Visit(func(key []byte, v *fastjson.Value) {
			keyName := string(key)
			// ignore the spcial fields
			if keyName == KeyMeasurement || keyName == KeyTimestamp {
				return
			}

			if v.Type() == fastjson.TypeString { // string -> tag
				// When a property name starts with a '$' character the user
				// wants to override the "strig = tag" association.
				// The Property is treated as a measurment field.
				if strings.HasPrefix(keyName, FieldOverride) {
					fields[strings.Trim(keyName, FieldOverride)] = v.String()
				} else {
					tags[keyName] = v.String()
				}
			} else if v.Type() == fastjson.TypeNumber { // number -> field value
				val, _ := v.Float64()
				fields[keyName] = val
			} else {
				p.log.Warnf("ignoring field \"%s\": invalid type %s",
					keyName, v.Type().String())
			}
		})

		// construct a new point and add to badge
		pt, err := influx.NewPoint(measurement, tags, fields, timestamp)
		if err != nil {
			p.log.Errorln("failed to construct new point:", err.Error())
			return
		}
		bt.AddPoint(pt)

	case fastjson.TypeArray:
		p.process(bt, v, time)
	}
}

// ---------------------------------------------------------------------------------------
//  private functions
// ---------------------------------------------------------------------------------------

// getTimestamp returns the timestamp set in the object when
// the special key is present. Otherwise it returs defaultTime.
func getTimestamp(obj *fastjson.Object, defaultTime time.Time) (time.Time, error) {
	t := obj.Get(KeyTimestamp)
	if t == nil {
		return defaultTime, nil
	}

	switch t.Type() {
	case fastjson.TypeNumber:
		// the timestamp field is a unix timestamp as integer in milliseconds
		unixMs, err := t.Int64()
		if err != nil {
			return defaultTime,
				fmt.Errorf("not using field \"%s\": %s", KeyTimestamp, err.Error())
		}

		// convert the milliseconds to nanoseconds and construct a time object
		return time.Unix(0, unixMs*1000.0*1000.0), nil
	default:
		return defaultTime,
			fmt.Errorf("not using field \"%s\": not a suitable type %s",
				KeyTimestamp, t.Type().String())
	}
}
