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
	"flag"
	"os"
	"syscall"

	"github.com/faryon93/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/faryon93/nsqflux/queue"
)

// ---------------------------------------------------------------------------------------
//  global variables
// ---------------------------------------------------------------------------------------

var (
	Color bool
	Conf  string
)

// ---------------------------------------------------------------------------------------
//  application entry
// ---------------------------------------------------------------------------------------

func main() {
	flag.BoolVar(&Color, "color", false, "force color logging")
	flag.StringVar(&Conf, "conf", "", "config path")
	flag.Parse()

	//setup logger
	formater := logrus.TextFormatter{ForceColors: Color}
	logrus.SetFormatter(&formater)
	logrus.SetOutput(os.Stdout)
	logrus.Infoln("starting", GetAppVersion())

	// parse the configruation file
	viper.SetConfigName("nsqflux")
	viper.AddConfigPath("/etc/nsqflux")
	viper.AddConfigPath(".")
	viper.SetConfigFile(Conf)
	viper.SetDefault("nsq.lookupd", "nsq_lookupd:4161")
	err := viper.ReadInConfig()
	if err != nil {
		logrus.Errorln("failed to read config:", err.Error())
		os.Exit(-1)
	}

	forwarders, err := GetForwarders()
	if err != nil {
		logrus.Errorln("failed to parse forwarders:", err.Error())
		os.Exit(-1)
	}

	// construct the queue listeners
	lookupd := queue.New(viper.GetString("nsq.lookupd"))
	defer lookupd.Stop()

	for i, forwarder := range forwarders {
		if err := forwarder.IsValid(); err != nil {
			logrus.Errorln("forwarder #%d is not valid: %s", i, err.Error())
			continue
		}

		logrus.Infof("forwarding \"%s\" to \"%s\"",
			forwarder.Topic, forwarder.Influx.Addr+"/"+forwarder.Influx.Database)

		handler := forwarder.Handle()
		if handler == nil {
			// error handling is already done in the generator function
			continue
		}

		// subscribe to the configured NSQ topic
		topic := forwarder.GetTopic()
		channel := forwarder.GetChannel()
		err := lookupd.Subscribe(topic, channel, handler)
		if err != nil {
			logrus.Errorf("failed to subscribe %s: %s",
				forwarder.Topic, err.Error())
			continue
		}
	}

	util.WaitSignal(os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	logrus.Infoln("received SIGINT / SIGTERM going to shutdown")
}
