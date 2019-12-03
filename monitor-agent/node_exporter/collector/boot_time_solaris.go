// Copyright 2018 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build solaris
// +build !noboottime

package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/siebenmann/go-kstat"
)

type bootTimeCollector struct {
	boottime typedDesc
}

func init() {
	registerCollector("boottime", defaultEnabled, newBootTimeCollector)
}

func newBootTimeCollector() (Collector, error) {
	return &bootTimeCollector{
		boottime: typedDesc{
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "", "boot_time_seconds"),
				"Unix time of last boot, including microseconds.",
				nil, nil,
			), prometheus.GaugeValue},
	}, nil
}

// newBootTimeCollector returns a new Collector exposing system boot time on Solaris systems.
// Update pushes boot time onto ch
func (c *bootTimeCollector) Update(ch chan<- prometheus.Metric) error {
	tok, err := kstat.Open()
	if err != nil {
		return err
	}

	defer tok.Close()

	ks, err := tok.Lookup("unix", 0, "system_misc")
	if err != nil {
		return err
	}

	v, err := ks.GetNamed("boot_time")
	if err != nil {
		return err
	}

	ch <- c.boottime.mustNewConstMetric(float64(v.UintVal))

	return nil
}
