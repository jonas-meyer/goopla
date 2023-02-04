package goopla

import "time"

const defaultStreamInterval = time.Minute * 1

type streamConfig struct {
	Interval       time.Duration
	DiscardInitial bool
}

type StreamOpt func(*streamConfig)

func StreamInterval(i time.Duration) StreamOpt {
	return func(c *streamConfig) {
		if i > 0 {
			c.Interval = i
		}
	}
}

func StreamDiscardInitial(c *streamConfig) {
	c.DiscardInitial = true
}
