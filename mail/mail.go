//
// Copyright (C) 2026 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.

package mail

import (
	"context"
	"log/slog"

	"github.com/wneessen/go-mail"
)

type Config interface {
}

type StaticConfig struct {
}

type PayloadFactory struct {
	client *mail.Client
	logger *slog.Logger
}

func NewPayloadFactory(config Config) (*PayloadFactory, error) {
	factory := &PayloadFactory{
		logger: slog.With(slog.String("transport", "Mail")),
	}
	return factory, nil
}

func (f *PayloadFactory) NewPlainPayload(body string) *Payload {
	payload := &Payload{
		factory: f,
	}
	return payload
}

type Payload struct {
	factory *PayloadFactory
}

func (p *Payload) Send(ctx context.Context, params any) error {
	return nil
}
