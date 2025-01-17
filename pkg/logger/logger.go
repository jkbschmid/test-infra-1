// Copyright 2019 Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logger

import (
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Log             logr.Logger
	configFromFlags = Config{}
)

var encoderConfig = zapcore.EncoderConfig{
	TimeKey:        "ts",
	LevelKey:       "level",
	NameKey:        "logger",
	CallerKey:      "caller",
	MessageKey:     "msg",
	StacktraceKey:  "stacktrace",
	LineEnding:     zapcore.DefaultLineEnding,
	EncodeLevel:    zapcore.LowercaseLevelEncoder,
	EncodeTime:     zapcore.ISO8601TimeEncoder,
	EncodeDuration: zapcore.SecondsDurationEncoder,
	EncodeCaller:   zapcore.ShortCallerEncoder,
}

var cliEncoderConfig = zapcore.EncoderConfig{
	TimeKey:        "",
	LevelKey:       "level",
	NameKey:        "logger",
	CallerKey:      "caller",
	MessageKey:     "msg",
	StacktraceKey:  "stacktrace",
	LineEnding:     zapcore.DefaultLineEnding,
	EncodeLevel:    zapcore.LowercaseColorLevelEncoder,
	EncodeTime:     zapcore.ISO8601TimeEncoder,
	EncodeDuration: zapcore.SecondsDurationEncoder,
	EncodeCaller:   zapcore.ShortCallerEncoder,
}

var defaultConfig = zap.Config{
	Level:             zap.NewAtomicLevelAt(zap.InfoLevel),
	Development:       true,
	Encoding:          "console",
	DisableStacktrace: false,
	DisableCaller:     false,
	EncoderConfig:     encoderConfig,
	OutputPaths:       []string{"stderr"},
	ErrorOutputPaths:  []string{"stderr"},
}

var productionConfig = zap.Config{
	Level:             zap.NewAtomicLevelAt(zap.InfoLevel),
	Development:       false,
	DisableStacktrace: true,
	DisableCaller:     true,
	Encoding:          "json",
	EncoderConfig:     encoderConfig,
	OutputPaths:       []string{"stderr"},
	ErrorOutputPaths:  []string{"stderr"},
}

func New(config *Config) (logr.Logger, error) {
	if config == nil {
		config = &configFromFlags
	}
	zapCfg := determineZapConfig(config)

	level := int8(0 - config.Verbosity)
	zapCfg.Level = zap.NewAtomicLevelAt(zapcore.Level(level))

	zapLog, err := zapCfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}
	Log = zapr.NewLogger(zapLog)
	return Log, nil
}

// NewCliLogger creates a new logger for cli usage.
// CLI usage means that by default:
// - the default dev config
// - encoding is console
// - timestamps are disabled (can be still activated by the cli flag)
// - level are color encoded
func NewCliLogger() (logr.Logger, error) {
	config := &configFromFlags
	config.Cli = true
	return New(config)
}

func determineZapConfig(loggerConfig *Config) zap.Config {
	var zapConfig zap.Config
	if loggerConfig.Development {
		zapConfig = defaultConfig
	} else if loggerConfig.Cli {
		zapConfig = defaultConfig
		zapConfig.EncoderConfig = cliEncoderConfig
	} else {
		zapConfig = productionConfig
	}

	loggerConfig.SetDisableCaller(&zapConfig)
	loggerConfig.SetDisableStacktrace(&zapConfig)
	loggerConfig.SetTimestamp(&zapConfig)

	return zapConfig
}
