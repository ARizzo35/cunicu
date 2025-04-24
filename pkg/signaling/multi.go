// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package signaling

import (
	"context"
	"net/url"

	"cunicu.li/cunicu/pkg/crypto"
	"cunicu.li/cunicu/pkg/log"
	signalingproto "cunicu.li/cunicu/pkg/proto/signaling"
	"go.uber.org/zap"
)

type MultiBackend struct {
	Backends []Backend
	logger   *log.Logger
}

func NewMultiBackend(uris []url.URL, cfg *BackendConfig) (*MultiBackend, error) {
	mb := &MultiBackend{
		Backends: []Backend{},
	}

	mb.logger = log.Global.Named("backend.multi")

	for _, u := range uris {
		cfg.URI = &u

		if b, err := NewBackend(cfg); err == nil {
			mb.Backends = append(mb.Backends, b)
		} else {
			return nil, err
		}
	}

	return mb, nil
}

func (mb *MultiBackend) Type() signalingproto.BackendType {
	return signalingproto.BackendType_MULTI
}

func (mb *MultiBackend) ByType(t signalingproto.BackendType) Backend {
	for _, b := range mb.Backends {
		if b.Type() == t {
			return b
		}
	}

	return nil
}

func (mb *MultiBackend) Publish(ctx context.Context, kp *crypto.KeyPair, msg *Message) error {
	var err error = nil
	var all_failed bool = true
	for _, b := range mb.Backends {
		if err = b.Publish(ctx, kp, msg); err != nil {
			mb.logger.Error("Failed to publish", zap.Error(err), zap.Any("type", b.Type()))
		} else {
			all_failed = false
		}
	}

	if all_failed {
		return err
	} else {
		return nil
	}
}

func (mb *MultiBackend) Subscribe(ctx context.Context, kp *crypto.KeyPair, h MessageHandler) (bool, error) {
	var err error = nil
	for _, b := range mb.Backends {
		if _, err = b.Subscribe(ctx, kp, h); err != nil {
			mb.logger.Error("Failed to subscribe", zap.Error(err), zap.Any("type", b.Type()))
		}
	}

	return false, err
}

func (mb *MultiBackend) Unsubscribe(ctx context.Context, kp *crypto.KeyPair, h MessageHandler) (bool, error) {
	var err error = nil
	for _, b := range mb.Backends {
		if _, err = b.Unsubscribe(ctx, kp, h); err != nil {
			mb.logger.Error("Failed to unsubscribe", zap.Error(err), zap.Any("type", b.Type()))
		}
	}

	return false, err
}

func (mb *MultiBackend) Close() error {
	var err error = nil
	for _, b := range mb.Backends {
		if err = b.Close(); err != nil {
			mb.logger.Error("Failed to close", zap.Error(err), zap.Any("type", b.Type()))
		}
	}

	return err
}
