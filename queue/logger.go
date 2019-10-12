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
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

// ---------------------------------------------------------------------------------------
//  types
// ---------------------------------------------------------------------------------------

type logger struct{}

// ---------------------------------------------------------------------------------------
//  public members
// ---------------------------------------------------------------------------------------

func (n *logger) Output(calldepth int, s string) error {
	if strings.HasPrefix(s, "INF") {
		s = strings.TrimSpace(strings.TrimPrefix(s, "INF"))
		s = strings.TrimSpace(strings.TrimPrefix(s, strconv.Itoa(calldepth)))
		logrus.Infoln(s)

	} else if strings.HasPrefix(s, "ERR") {
		s = strings.TrimSpace(strings.TrimPrefix(s, "ERR"))
		s = strings.TrimSpace(strings.TrimPrefix(s, strconv.Itoa(calldepth)))
		logrus.Errorln(s)

	} else if strings.HasPrefix(s, "WRN") {
		s = strings.TrimSpace(strings.TrimPrefix(s, "WRN"))
		s = strings.TrimSpace(strings.TrimPrefix(s, strconv.Itoa(calldepth)))
		logrus.Warnln(s)

	} else {
		logrus.Errorln(s)
	}

	return nil
}
