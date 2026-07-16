//
// Copyright (C) 2026 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.

package notify

import (
	"context"
	"log/slog"
)

type LogPayload struct {
	Logger *slog.Logger
	Level  slog.Level
	Msg    string
}

func (p *LogPayload) Send(ctx context.Context, params any) error {
	namedValues, err := DecodeParams(params)
	if err != nil {
		return err
	}
	args := make([]any, 0, len(namedValues))
	for _, namedValue := range namedValues {
		args = append(args, slog.Any(namedValue.Name, namedValue.Value))
	}
	logger := p.Logger
	if logger == nil {
		logger = slog.Default()
	}
	logger.Log(ctx, p.Level, p.Msg, args...)
	return nil
}
