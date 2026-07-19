//
// Copyright (C) 2026 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.

package mattermost

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/tdrn-org/go-notify"
)

type ChannelResolver[T any] interface {
	ResolveChannelID(ctx context.Context, client *model.Client4, params T) (string, error)
}

type Config[T any] interface {
	GetServerURL() (string, error)
	GetToken() (string, error)
	ChannelResolver[T]
}

type StaticConfig struct {
	ServerURL string
	Token     string
	ChannelID string
}

func (c *StaticConfig) GetServerURL() (string, error) {
	return c.ServerURL, nil
}

func (c *StaticConfig) GetToken() (string, error) {
	return c.Token, nil
}

func (c *StaticConfig) ResolveChannelID(_ context.Context, _ *model.Client4, _ any) (string, error) {
	return c.ChannelID, nil
}

type PayloadFactory[T any] struct {
	client          *model.Client4
	channelResolver ChannelResolver[T]
	logger          *slog.Logger
}

func NewPayloadFactory[T any](config Config[T]) (*PayloadFactory[T], error) {
	serverURL, err := config.GetServerURL()
	if err != nil {
		return nil, err
	}
	token, err := config.GetToken()
	if err != nil {
		return nil, err
	}
	client := model.NewAPIv4Client(serverURL)
	client.SetToken(token)
	payloadFactory := &PayloadFactory[T]{
		client:          client,
		channelResolver: config,
		logger:          slog.With(slog.String("transport", "Mattermost"), slog.String("server", serverURL)),
	}
	return payloadFactory, nil
}

func (f *PayloadFactory[T]) NewPayload(message string) *Payload[T] {
	payload := &Payload[T]{
		factory: f,
		message: message,
	}
	return payload
}

type Payload[T any] struct {
	factory *PayloadFactory[T]
	message string
}

func (p *Payload[T]) Send(ctx context.Context, params any) error {
	mattermostParams, ok := params.(T)
	if !ok {
		return fmt.Errorf("unexpected params type: %s", reflect.TypeOf(params))
	}
	channelID, err := p.factory.channelResolver.ResolveChannelID(ctx, p.factory.client, mattermostParams)
	if err != nil {
		return fmt.Errorf("failed to resolve Mattermost channel (cause: %w)", err)
	}
	message, err := notify.ExecuteTextTemplate(p.message, params)
	if err != nil {
		return fmt.Errorf("failed to prepare Mattermost message (cause: %w)", err)
	}
	post := &model.Post{
		ChannelId: channelID,
		Message:   message,
	}
	p.factory.logger.Info("creating Mattermost post...")
	createdPost, response, err := p.factory.client.CreatePost(ctx, post)
	if err != nil {
		return fmt.Errorf("failed to send Mattermost post (cause: %w)", err)
	}
	if response.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create Mattermost post (status code: %d)", response.StatusCode)
	}
	p.factory.logger.Debug("Mattermost post created", slog.String("id", createdPost.Id))
	return nil
}
