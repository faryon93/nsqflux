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
	_ "github.com/influxdata/influxdb1-client"
	influx "github.com/influxdata/influxdb1-client/v2"
)

// ---------------------------------------------------------------------------------------
//  types
// ---------------------------------------------------------------------------------------

type Influx struct {
	Addr        string `yaml:"addr" json:"addr"`
	User        string `yaml:"user" json:"user"`
	Password    string `yaml:"password" json:"password"`
	Database    string `yaml:"database" json:"database"`
	Measurement string `yaml:"measurement" json:"measurement"`
}

func (i *Influx) GetClient() (influx.Client, error) {
	return influx.NewHTTPClient(i.toHTTPConfig())
}

// ---------------------------------------------------------------------------------------
// private functions
// ---------------------------------------------------------------------------------------

func (i *Influx) toHTTPConfig() influx.HTTPConfig {
	return influx.HTTPConfig{
		Addr:     i.Addr,
		Username: i.User,
		Password: i.Password,
	}
}
