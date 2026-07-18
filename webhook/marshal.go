//
// Copyright (C) 2026 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.

package webhook

import (
	"encoding/json"
	"fmt"
	"strings"
)

type bodyMarshaler interface {
	ContentType() string
	Marshal(body any) (string, error)
}

type jsonBodyMarshaler struct{}

func (m *jsonBodyMarshaler) ContentType() string {
	return "application/json"
}

func (m *jsonBodyMarshaler) Marshal(body any) (string, error) {
	buffer := &strings.Builder{}
	err := json.NewEncoder(buffer).Encode(body)
	if err != nil {
		return "", fmt.Errorf("failed to encode JSON body (cause: %w)", err)
	}
	return buffer.String(), nil
}
