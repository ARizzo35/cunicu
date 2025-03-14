// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cunicu.li/cunicu/pkg/backoff"
	"cunicu.li/cunicu/pkg/crypto"
	"cunicu.li/cunicu/pkg/log"
	"cunicu.li/cunicu/pkg/proto"
	signalingproto "cunicu.li/cunicu/pkg/proto/signaling"
	"cunicu.li/cunicu/pkg/signaling"
)

func init() { //nolint:gochecknoinits
	signaling.Backends["grpc"] = &signaling.BackendPlugin{
		New:         NewBackend,
		Description: "gRPC",
	}
}

type Backend struct {
	signaling.SubscriptionsRegistry

	client signalingproto.SignalingClient
	conn   *grpc.ClientConn
	//connected bool

	config BackendConfig

	logger *log.Logger
}

func NewBackend(cfg *signaling.BackendConfig, logger *log.Logger) (signaling.Backend, error) {
	var err error

	b := &Backend{
		SubscriptionsRegistry: signaling.NewSubscriptionsRegistry(),
		logger:                logger,
		//connected:             false,
	}

	if err := b.config.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse backend configuration: %w", err)
	}

	if b.conn, err = grpc.NewClient(b.config.Target, b.config.Options...); err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	b.client = signalingproto.NewSignalingClient(b.conn)

	go func() {
		bo := &backoff.ExponentialBackOff{
			InitialInterval:     500 * time.Millisecond,
			RandomizationFactor: 0.5,
			Multiplier:          1.5,
			MaxInterval:         1 * time.Minute,
		}
		for _, d := range backoff.Retry(bo) {
			if bi, err := b.client.GetBuildInfo(context.Background(), &proto.Empty{}, grpc.WaitForReady(false)); err != nil {
				b.logger.Error("Failed to get build info from the gRPC signaling server", zap.Error(err), zap.Duration("after", d))
			} else {
				//b.connected = true

				b.logger.Debug("Connected to GRPC signaling server",
					zap.String("server_arch", bi.Arch),
					zap.String("server_version", bi.Version),
					zap.String("server_commit", bi.Commit),
					zap.String("server_tag", bi.Tag),
					zap.String("server_branch", bi.Branch),
					zap.String("server_os", bi.Os),
				)

				for _, h := range cfg.OnReady {
					h.OnSignalingBackendReady(b)
				}

				break
			}
		}
	}()

	return b, nil
}

func (b *Backend) Type() signalingproto.BackendType {
	return signalingproto.BackendType_GRPC
}

func (b *Backend) Subscribe(ctx context.Context, kp *crypto.KeyPair, h signaling.MessageHandler) (bool, error) {
	first, err := b.SubscriptionsRegistry.Subscribe(kp, h)
	if err != nil {
		return false, err
	} else if first {
		pk := kp.Ours.PublicKey()

		return first, b.subscribeFromServer(ctx, &pk)
	}

	return first, nil
}

// Unsubscribe from messages send by a specific peer.
func (b *Backend) Unsubscribe(ctx context.Context, kp *crypto.KeyPair, h signaling.MessageHandler) (bool, error) {
	last, err := b.SubscriptionsRegistry.Unsubscribe(kp, h)
	if err != nil {
		return false, err
	} else if last {
		pk := kp.Ours.PublicKey()

		return last, b.unsubscribeFromServer(ctx, &pk)
	}

	return last, nil
}

func (b *Backend) Publish(ctx context.Context, kp *crypto.KeyPair, msg *signaling.Message) error {
	//if !b.connected {
	//	return signaling.ErrNotReady
	//}

	env, err := msg.Encrypt(kp)
	if err != nil {
		return fmt.Errorf("failed to encrypt message: %w", err)
	}

	if _, err = b.client.Publish(ctx, env, grpc.WaitForReady(false)); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

func (b *Backend) Close() error {
	if err := b.conn.Close(); err != nil {
		return fmt.Errorf("failed to close gRPC connection: %w", err)
	}

	return nil
}

func (b *Backend) subscribeFromServer(ctx context.Context, pk *crypto.Key) error {
	params := &signalingproto.SubscribeParams{
		Key: pk.Bytes(),
	}

	go func() {
		bo := &backoff.ExponentialBackOff{
			InitialInterval:     500 * time.Millisecond,
			RandomizationFactor: 0.5,
			Multiplier:          1.5,
			MaxInterval:         1 * time.Minute,
		}
	outer:
		for _, d := range backoff.Retry(bo) {
			stream, err := b.client.Subscribe(ctx, params, grpc.WaitForReady(false))
			if err != nil {
				b.logger.Error("failed to subscribe to offers", zap.Error(err), zap.Duration("after", d))
				continue
			}
			// Wait until subscription has been created
			// This avoids a race between Subscribe() / Publish() when two subscribers are subscribing
			// to each other.
			if _, err := stream.Recv(); err != nil {
				b.logger.Error("failed receive synchronization envelope", zap.Error(err), zap.Duration("after", d))
				continue
			}

			b.logger.Debug("Created new subscription", zap.Any("pk", pk))

			for {
				env, err := stream.Recv()
				if err != nil {
					if !errors.Is(err, io.EOF) && status.Code(err) != codes.Canceled {
						b.logger.Error("Subscription stream closed. Re-subscribing..", zap.Error(err))
						continue outer
					}

					break outer
				}

				if err := b.SubscriptionsRegistry.NewMessage(env); err != nil {
					b.logger.Error("Failed to decrypt message", zap.Error(err))
				}
			}
		}
	}()

	return nil
}

func (b *Backend) unsubscribeFromServer(_ context.Context, _ *crypto.Key) error {
	// TODO: Cancel subscription stream
	return nil
}
