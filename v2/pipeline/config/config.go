package config

import (
	"pack.ag/amqp"
	"time"
)

type Config struct {
	Session *amqp.Session
	Client  *amqp.Client
	WSize 	uint32
	Timeout time.Duration
}