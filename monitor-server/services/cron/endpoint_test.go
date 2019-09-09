package cron

import "testing"

func TestGetEndpointData(t *testing.T) {
	GetEndpointData("192.168.67.131", "9100", "node_")
}