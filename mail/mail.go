//
// Copyright (C) 2026 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.

package mail

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	stdmail "net/mail"
	"strconv"

	"github.com/tdrn-org/go-notify"
	"github.com/tdrn-org/go-pool"
	"github.com/wneessen/go-mail"
)

type RecipientsResolver[T any] interface {
	ResolveRecipients(ctx context.Context, params T) ([]*stdmail.Address, error)
}

type SubjectResolver[T any] interface {
	ResolveSubject(ctx context.Context, params T) (string, error)
}

type Config[T any] interface {
	GetServerAddress() (string, int, error)
	GetUser() (string, error)
	GetPassword() (string, error)
	GetFromAddress() (string, error)
	GetFromName() (string, error)
	RecipientsResolver[T]
	SubjectResolver[T]
}

type StaticConfig struct {
	ServerAddress string
	User          string
	Password      string
	FromAddress   string
	FromName      string
	ToAddress     string
	ToName        string
	Subject       string
}

func (c *StaticConfig) GetServerAddress() (string, int, error) {
	host, portString, err := net.SplitHostPort(c.ServerAddress)
	port := 0
	if err == nil {
		port, err = strconv.Atoi(portString)
		if err != nil {
			return "", 0, fmt.Errorf("invalid port number in Mail server address '%s' (cause: %w)", c.ServerAddress, err)
		}
	} else {
		host = c.ServerAddress
	}
	return host, port, nil
}

func (c *StaticConfig) GetUser() (string, error) {
	return c.User, nil
}

func (c *StaticConfig) GetPassword() (string, error) {
	return c.Password, nil
}

func (c *StaticConfig) GetFromAddress() (string, error) {
	return c.FromAddress, nil
}

func (c *StaticConfig) GetFromName() (string, error) {
	return c.FromName, nil
}

func (c *StaticConfig) ResolveRecipients(_ context.Context, _ any) ([]*stdmail.Address, error) {
	address := c.ToAddress
	if c.ToName != "" {
		address = fmt.Sprintf("%s <%s>", c.ToName, c.ToName)
	}
	resolved, err := stdmail.ParseAddress(address)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Mail address '%s' (cause: %w)", address, err)
	}
	return []*stdmail.Address{resolved}, nil
}

func (c *StaticConfig) ResolveSubject(_ context.Context, _ any) (string, error) {
	return c.Subject, nil
}

type PayloadFactory[T any] struct {
	clientPool         *pool.Resources[*mail.Client]
	fromAddress        string
	fromName           string
	recipientsResolver RecipientsResolver[T]
	subjectResolver    SubjectResolver[T]
	logger             *slog.Logger
}

func NewPayloadFactory[T any](config Config[T]) (*PayloadFactory[T], error) {
	host, port, err := config.GetServerAddress()
	if err != nil {
		return nil, err
	}
	user, err := config.GetUser()
	if err != nil {
		return nil, err
	}
	password, err := config.GetPassword()
	if err != nil {
		return nil, err
	}
	mailClientFactory := &mailClientFactory{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
	}
	fromAddress, err := config.GetFromAddress()
	if err != nil {
		return nil, err
	}
	fromName, err := config.GetFromName()
	if err != nil {
		return nil, err
	}
	factory := &PayloadFactory[T]{
		clientPool:         pool.NewResourcePool("notify", pool.ResourceFactory[*mail.Client](mailClientFactory)),
		fromAddress:        fromAddress,
		fromName:           fromName,
		recipientsResolver: config,
		subjectResolver:    config,
		logger:             slog.With(slog.String("transport", "Mail")),
	}
	return factory, nil
}

func (f *PayloadFactory[T]) NewPlainPayload(body string) *Payload[T] {
	payload := &Payload[T]{
		factory:     f,
		contentType: mail.TypeTextPlain,
		body:        body,
	}
	return payload
}

func (f *PayloadFactory[T]) Shutdown(ctx context.Context) error {
	return f.clientPool.Shutdown(ctx)
}

func (f *PayloadFactory[T]) Close() error {
	return f.clientPool.Close()
}

type mailClientFactory struct {
	Host     string
	Port     int
	User     string
	Password string
}

func (f *mailClientFactory) New(ctx context.Context) (*mail.Client, error) {
	options := make([]mail.Option, 0)
	if f.User != "" {
		options = append(options, mail.WithUsername(f.User), mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover))
		if f.Password != "" {
			options = append(options, mail.WithPassword(f.Password))
		}
	}
	if f.Port != 0 {
		options = append(options, mail.WithPort(f.Port))
	}
	client, err := mail.NewClient(f.Host, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Mail client (cause: %w)", err)
	}
	return client, nil
}

type Payload[T any] struct {
	factory     *PayloadFactory[T]
	contentType mail.ContentType
	body        string
}

func (p *Payload[T]) Send(ctx context.Context, params T) error {
	message, err := p.prepareMessage(ctx, params)
	if err != nil {
		return err
	}
	client, err := p.factory.clientPool.Get(ctx)
	if err != nil {
		return err
	}
	defer client.Release()
	return p.sendMessage(ctx, client.Get(), message)
}

func (p *Payload[T]) sendMessage(ctx context.Context, client *mail.Client, message *mail.Msg) error {
	logger := p.factory.logger.With(slog.Any("to", message.GetToString()))
	logger.Info("sending Mail message...")
	err := client.DialAndSendWithContext(ctx, message)
	if err != nil {
		logger.Warn("failed to send Mail message; retry after reset (cause: %w)", slog.Any("err", err))
		err := client.Reset()
		if err != nil {
			return fmt.Errorf("failed to reset Mail for re-seend attempt (cause: %w)", err)
		}
		err = client.DialAndSendWithContext(ctx, message)
		if err != nil {
			return fmt.Errorf("failed to send Mail (cause: %w)", err)
		}
	}
	logger.Info("Mail message sent")
	return nil
}

func (p *Payload[T]) prepareMessage(ctx context.Context, params T) (*mail.Msg, error) {
	message := mail.NewMsg()
	err := p.prepareMessageFrom(ctx, message, params)
	if err != nil {
		return nil, err
	}
	err = p.prepareMessageTo(ctx, message, params)
	if err != nil {
		return nil, err
	}
	err = p.prepareMessageSubject(ctx, message, params)
	if err != nil {
		return nil, err
	}
	body := p.body
	switch p.contentType {
	case mail.TypeTextPlain:
		body, err = notify.ExecuteTextTemplate(body, params)
	case mail.TypeTextHTML:
		body, err = notify.ExecuteHTMLTemplate(body, params)
	default:
		return nil, fmt.Errorf("unexpected Mail content type: '%s'", p.contentType)
	}
	if err != nil {
		return nil, err
	}
	message.SetBodyString(p.contentType, body)
	return message, nil
}

func (p *Payload[T]) prepareMessageFrom(_ context.Context, message *mail.Msg, _ T) error {
	var err error
	if p.factory.fromName == "" {
		err = message.From(p.factory.fromAddress)
	} else {
		err = message.FromFormat(p.factory.fromName, p.factory.fromAddress)
	}
	if err != nil {
		return fmt.Errorf("failed to set Mail from address (cause: %w)", err)
	}
	return nil
}

func (p *Payload[T]) prepareMessageTo(ctx context.Context, message *mail.Msg, params T) error {
	recipients, err := p.factory.recipientsResolver.ResolveRecipients(ctx, params)
	if err != nil {
		return err
	}
	for _, recipient := range recipients {
		message.AddToMailAddress(recipient)
	}
	return nil
}

func (p *Payload[T]) prepareMessageSubject(ctx context.Context, message *mail.Msg, params T) error {
	subject, err := p.factory.subjectResolver.ResolveSubject(ctx, params)
	if err != nil {
		return err
	}
	message.Subject(subject)
	return nil
}
