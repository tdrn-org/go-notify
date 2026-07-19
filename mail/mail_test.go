//
// Copyright (C) 2026 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.

package mail_test

import (
	"testing"

	"github.com/go-openapi/testify/require"
	"github.com/tdrn-org/go-notify/mail"
)

func TestMail(t *testing.T) {
	config := newMailConfig()
	if config == nil {
		t.SkipNow()
	}
	factory, err := mail.NewPayloadFactory(config)
	require.NoError(t, err)
	body := "Hello world from {{.Testname}}!"
	payload := factory.NewPlainPayload(body)
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

func newMailConfig() mail.Config[*Params] {
	return nil
}
