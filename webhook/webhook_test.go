//
// Copyright (C) 2026 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.

package webhook_test

import (
	"testing"

	"github.com/go-openapi/testify/require"
	"github.com/tdrn-org/go-notify/webhook"
)

func TestWebook(t *testing.T) {
	config := newWebhookConfig()
	if config == nil {
		t.SkipNow()
	}
	factory, err := webhook.NewPayloadFactory(config)
	require.NoError(t, err)
	body := &Body{
		Text: "Hello world from {{.Testname}}!",
	}
	payload := factory.NewJSONPayload(body, nil)
	params := &Params{
		Testname: t.Name(),
	}
	err = payload.Send(t.Context(), params)
	require.NoError(t, err)
}

type Body struct {
	Text string `json:"text"`
}

type Params struct {
	Testname string
}

func newWebhookConfig() webhook.Config {
	return nil
}
