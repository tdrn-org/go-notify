//
// Copyright (C) 2026 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.

package notify_test

import (
	"testing"

	"github.com/go-openapi/testify/require"
	"github.com/tdrn-org/go-notify"
)

func TestLogPayload(t *testing.T) {
	payload := &notify.LogPayload{Msg: t.Name()}
	params := &LogParams{
		Name1: "value1",
		Name2: true,
	}
	err := payload.Send(t.Context(), params)
	require.NoError(t, err)
}

type LogParams struct {
	Name1 string `name:"name1"`
	Name2 bool   `name:"name2"`
}
