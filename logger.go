package main

import (
	appsutils "github.com/mattermost/mattermost-plugin-apps/utils"
	"go.uber.org/zap/zapcore"
)

var logger = appsutils.MustMakeCommandLogger(zapcore.InfoLevel)
