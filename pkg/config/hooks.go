// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/pion/ice/v4"
	"github.com/pion/stun/v3"

	icex "cunicu.li/cunicu/pkg/ice"
)

var errUnknownHookType = errors.New("unknown hook type")

func hookDecodeHook(f, t reflect.Type, data any) (any, error) {
	if f.Kind() != reflect.Map {
		return data, nil
	}

	if t.Name() != "HookSetting" {
		return data, nil
	}

	var base BaseHookSetting
	if err := mapstructure.Decode(data, &base); err != nil { //nolint:musttag
		return nil, err
	}

	var hook HookSetting

	switch base.Type {
	case "web":
		hook = &WebHookSetting{
			Method: "POST",
		}
	case "exec":
		hook = &ExecHookSetting{
			Stdin: true,
		}
	default:
		return nil, fmt.Errorf("%w: %s", errUnknownHookType, base.Type)
	}

	decoder, err := mapstructure.NewDecoder(DecoderConfig(hook))
	if err != nil {
		return nil, err
	}

	return hook, decoder.Decode(data)
}

// stringsDecodeHook is a DecodeHookFunc that converts strings to various types.
func stringsDecodeHook(
	_ reflect.Type,
	t reflect.Type,
	data interface{},
) (interface{}, error) {
	str, ok := data.(string)
	if !ok {
		return data, nil
	}

	switch t {
	case reflect.TypeOf(stun.URI{}):
		u, err := stun.ParseURI(str)
		if err != nil {
			return nil, err
		}

		return *u, nil

	case reflect.TypeOf(url.URL{}):
		if !strings.Contains(str, ":") {
			str += ":"
		}

		u, err := url.Parse(str)
		if err != nil {
			return nil, err
		}

		return *u, nil

	case reflect.TypeOf(net.IPAddr{}):
		ip, err := net.ResolveIPAddr("ip", str)

		return *ip, err

	case reflect.TypeOf(net.IPNet{}):
		ip, net, err := net.ParseCIDR(str)
		if err != nil {
			return nil, err
		}

		net.IP = ip

		return net, nil

	case reflect.TypeOf(ice.NetworkTypeTCP4):
		return icex.ParseNetworkType(str)

	case reflect.TypeOf(ice.CandidateTypeUnspecified):
		return icex.ParseCandidateType(str)

	default:
		return data, nil
	}
}
