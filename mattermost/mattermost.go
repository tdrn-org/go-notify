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

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/tdrn-org/go-notify"
)

type Config interface {
	GetServerURL() (string, error)
	GetToken() (string, error)
	GetChannelID(client *model.Client4) (string, error)
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

func (c *StaticConfig) GetChannelID(_ *model.Client4) (string, error) {
	return c.ChannelID, nil
}

type PayloadFactory struct {
	client    *model.Client4
	channelID string
	logger    *slog.Logger
}

func NewPayloadFactory(config Config) (*PayloadFactory, error) {
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
	channelID, err := config.GetChannelID(client)
	if err != nil {
		return nil, err
	}
	payloadFactory := &PayloadFactory{
		client:    client,
		channelID: channelID,
		logger:    slog.With(slog.String("transport", "Mattermost"), slog.String("server", serverURL)),
	}
	return payloadFactory, nil
}

func (f *PayloadFactory) NewPayload(post *Post) *Payload {
	payload := &Payload{
		factory: f,
		post:    post,
	}
	return payload
}

type Post struct {
	Message string
}

type Payload struct {
	factory *PayloadFactory
	post    *Post
}

func (p *Payload) Send(ctx context.Context, params any) error {
	message, err := notify.ExecuteTextTemplate(p.post.Message, params)
	if err != nil {
		return fmt.Errorf("failed to prepare Mattermost message (cause: %w)", err)
	}
	post := &model.Post{
		ChannelId: p.factory.channelID,
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
