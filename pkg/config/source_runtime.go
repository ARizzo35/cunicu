// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stv0g/cunicu/pkg/log"
	"go.uber.org/zap"
)

type runtimeSource struct {
	*LocalFileProvider
	*koanf.Koanf

	watchCallback func(event any, err error)
}

func newRuntimeSource() *runtimeSource {
	source := &runtimeSource{
		LocalFileProvider: NewLocalFileProvider(RuntimeConfigFile),
		Koanf:             koanf.New("."),
	}

	return source
}

func (s *runtimeSource) Config() *koanf.Koanf {
	return s.Koanf
}

func (s *runtimeSource) Load() error {
	parser := yaml.Parser()
	if err := s.Koanf.Load(s.LocalFileProvider, parser); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		} else if errors.Is(err, os.ErrPermission) {
			log.Global.Warn("Failed to load runtime configuration", zap.Error(err))
			return nil
		}

		return err
	}

	return nil
}

func (s *runtimeSource) Watch(cb func(event any, err error)) error {
	if !hasRuntimeConfig() {
		s.watchCallback = cb
		return nil
	}

	return s.LocalFileProvider.Watch(cb)
}

// SaveRuntime saves the current runtime configuration to disk
func (s *runtimeSource) Save() error {
	f, err := os.OpenFile(RuntimeConfigFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}

	defer f.Close()

	fmt.Fprintln(f, "# This is the cunīcu runtime configuration file.")
	fmt.Fprintln(f, "# It contains configuration adjustments made by")
	fmt.Fprintln(f, "# by the user with the cunicu-config-set(1) command.")
	fmt.Fprintln(f, "#")
	fmt.Fprintln(f, "# Please do not edit this file by hand as it will")
	fmt.Fprintln(f, "# be overwritten by cunīcu.")
	fmt.Fprintln(f, "#")
	fmt.Fprintf(f, "# Last modification at %s\n", time.Now().Format(time.RFC1123Z))
	fmt.Fprintln(f, "---")

	if s.watchCallback != nil {
		if s.LocalFileProvider.Watch(s.watchCallback); err != nil {
			return err
		}
	}

	return s.Marshal(f)
}

func (s *runtimeSource) Marshal(wr io.Writer) error {
	return marshal(s.Koanf, wr)
}

// Update sets multiple settings in the provided map.
func (s *runtimeSource) Update(sets map[string]any) error {
	newKoanf := koanf.New(".")

	if err := newKoanf.Load(confmap.Provider(sets, "."), nil); err != nil {
		return err
	}

	// Check if settings are valid
	if _, err := unmarshal(newKoanf); err != nil {
		return err
	}

	s.Koanf = newKoanf

	return nil
}

func hasRuntimeConfig() bool {
	fi, err := os.Stat(RuntimeConfigFile)
	return err == nil && !fi.IsDir()
}
