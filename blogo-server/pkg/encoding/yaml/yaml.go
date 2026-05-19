// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package yaml

import "gopkg.in/yaml.v3"

var (
	Marshal    = yaml.Marshal
	Unmarshal  = yaml.Unmarshal
	NewDecoder = yaml.NewDecoder
	NewEncoder = yaml.NewEncoder
)
