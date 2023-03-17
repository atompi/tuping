package options

import (
	"strconv"
	"time"

	"github.com/spf13/viper"
)

var Version string = "1.0.0"

type PingOptions struct {
	Host     string
	Port     int
	Protocol string
	Count    int
	Size     int
	Interval time.Duration
	TTL      int
	Wait     bool
	DNS      string
}

func NewPingOptions(args []string) *PingOptions {
	var s PingOptions
	if len(args) < 2 {
		s.Host = args[0]
		s.Port = -1
	} else {
		s.Host = args[0]
		s.Port, _ = strconv.Atoi(args[1])
	}
	s.Protocol = viper.GetString("protocol")
	s.Count = viper.GetInt("count")
	s.Size = viper.GetInt("size")
	s.Interval = viper.GetDuration("interval")
	s.TTL = viper.GetInt("ttl")
	s.Wait = viper.GetBool("wait")
	s.DNS = viper.GetString("dns")
	return &s
}
