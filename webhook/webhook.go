//
// Copyright (C) 2026 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.

package webhook

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tdrn-org/go-notify"
)

type Config interface {
	GetClient() (*http.Client, error)
	GetWebhookURL() (string, error)
	GetTimeout() (time.Duration, error)
}

type StaticConfig struct {
	Client     *http.Client
	WebhookURL string
	Timeout    time.Duration
}

func (c *StaticConfig) GetClient() (*http.Client, error) {
	return c.Client, nil
}

func (c *StaticConfig) GetWebhookURL() (string, error) {
	return c.WebhookURL, nil
}

func (c *StaticConfig) GetTimeout() (time.Duration, error) {
	return c.Timeout, nil
}

type PayloadFactory struct {
	client     *http.Client
	webhookURL string
	timeout    time.Duration
	logger     *slog.Logger
}

func NewPayloadFactory(config Config) (*PayloadFactory, error) {
	client, err := config.GetClient()
	if err != nil {
		return nil, err
	}
	webhookURL, err := config.GetWebhookURL()
	if err != nil {
		return nil, err
	}
	parsedWebhookURL, err := url.Parse(webhookURL)
	if err != nil {
		return nil, fmt.Errorf("invalid or empty Webhook URL (cause: %w)", err)
	}
	timeout, err := config.GetTimeout()
	if err != nil {
		return nil, err
	}
	if client == nil {
		client = &http.Client{
			Timeout: timeout,
		}
	}
	factory := &PayloadFactory{
		client:     client,
		webhookURL: webhookURL,
		timeout:    timeout,
		logger:     slog.With(slog.String("transport", "Webhook"), slog.String("url", fmt.Sprintf("%s://%s", parsedWebhookURL.Scheme, parsedWebhookURL.Host))),
	}
	return factory, nil
}

type Headers [][2]string

func (f *PayloadFactory) NewJSONPayload(body any, headers Headers) *Payload {
	payload := &Payload{
		factory:   f,
		marshaler: &jsonBodyMarshaler{},
		body:      body,
		headers:   headers,
	}
	return payload
}

type Payload struct {
	factory   *PayloadFactory
	marshaler bodyMarshaler
	body      any
	headers   Headers
}

func (p *Payload) Send(ctx context.Context, params any) error {
	body, err := p.marshaler.Marshal(p.body)
	if err != nil {
		return fmt.Errorf("failed to marshal Webhook body (cause: %w)", err)
	}
	body, err = notify.ExecuteTextTemplate(body, params)
	if err != nil {
		return fmt.Errorf("failed to prepare Webhook body (cause: %w)", err)
	}
	req, err := http.NewRequest(http.MethodPost, p.factory.webhookURL, strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create Webhook request (cause: %w)", err)
	}
	req.Header.Set("Content-Type", p.marshaler.ContentType())
	for _, header := range p.headers {
		req.Header.Set(header[0], header[1])
	}
	p.factory.logger.Info("invoking Webhook...")
	rsp, err := p.factory.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to process Webhook process (cause: %w)", err)
	}
	if rsp.StatusCode < 200 || 299 < rsp.StatusCode {
		return fmt.Errorf("failed to invoke Webhook (status: %s)", rsp.Status)
	}
	p.factory.logger.Debug("Webhook invoked", slog.String("status", rsp.Status))
	return nil
}
