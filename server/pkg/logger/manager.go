package logger

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/rs/xid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var globalLogger *zap.Logger

type LoggerManager interface {
	InitLogger() error
	InfoWithElapsedTime(ctx context.Context, action string, startTime time.Time, optionalFields ...map[string]interface{})
	ErrorWithElapsedTime(ctx context.Context, action string, startTime time.Time, err error, optionalFields ...map[string]interface{})
}

type Manager struct {
	AppName       string
	Env           string
	LogFolderPath string
}

func NewManager(appName, env, logFolderPath string) (*Manager, error) {
	return &Manager{
		AppName:       appName,
		Env:           env,
		LogFolderPath: fmt.Sprintf("%s/%s", logFolderPath, appName),
	}, nil
}

func (m *Manager) generateLogFilePath() string {
	envMap := map[string]string{
		"development":        "dev",
		"docker-development": "docker_dev",
		"test":               "test",
		"production":         "prod",
	}
	env, ok := envMap[m.Env]
	if !ok {
		env = ""
	}

	date := time.Now().Format("20060102")
	logFileName := fmt.Sprintf("%s_%s_%s.log", m.AppName, env, date)
	logFilePath := fmt.Sprintf("%s/%s", m.LogFolderPath, logFileName)

	// Create the folder if it doesn't exist
	if err := os.MkdirAll(m.LogFolderPath, 0755); err != nil {
		fmt.Printf("failed to create log folder: %v\n", err)
	}

	return logFilePath
}

func (m *Manager) InitLogger() error {
	// Set Encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Set Writers
	logFilePath := m.generateLogFilePath()

	file, _ := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	consoleSyncer := zapcore.AddSync(os.Stdout)
	fileSyncer := zapcore.AddSync(file)

	// Set Core
	core := zapcore.NewTee(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig), // Use JSON
			consoleSyncer,                         // Output to console
			zapcore.InfoLevel,                     // Log level
		),
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig), // Use JSON
			fileSyncer,                            // Output to file
			zapcore.InfoLevel,                     // Log level
		),
	)

	logger := zap.New(core)
	logger.Info("Logger initialized", zap.String("log_file", logFilePath))

	globalLogger = logger
	return nil
}

// Convert to the format required by the logger and sort it, putting log_id first
func getSortedZapField(fields map[string]interface{}) []zapcore.Field {
	zapFields := make([]zapcore.Field, 0, len(fields))
	if logID, exists := fields["log_id"]; exists {
		zapFields = append(zapFields, zap.Any("log_id", logID))
		delete(fields, "log_id") // Remove log_id from the map to avoid adding it again
	}

	// Sort and add the remaining fields
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		zapFields = append(zapFields, zap.Any(k, fields[k]))
	}

	return zapFields
}

func getUserIdByCtx(ctx context.Context) (string, error) {
	userId := ctx.Value("userId")

	userIdStr, ok := userId.(string)
	if !ok && userId != nil {
		return "", errors.New("failed to convert value to string")
	}

	return userIdStr, nil
}

func logWithElapsedTime(ctx context.Context, startTime time.Time, optionalFields ...map[string]interface{}) []zapcore.Field {
	logID := xid.New().String()

	fields := make(map[string]interface{})
	fields["log_id"] = logID

	if len(optionalFields) > 0 {
		for k, v := range optionalFields[0] {
			fields[k] = v
		}
	}

	requester, err := getUserIdByCtx(ctx)
	if err != nil {
		globalLogger.Error(err.Error())
	}

	fields["requester"] = requester
	fields["elapsed_time"] = time.Since(startTime)

	return getSortedZapField(fields)
}

func (m *Manager) InfoWithElapsedTime(ctx context.Context, action string, startTime time.Time, optionalFields ...map[string]interface{}) {
	zapFields := logWithElapsedTime(ctx, startTime, optionalFields...)
	globalLogger.Info(action, zapFields...)
}

func (m *Manager) ErrorWithElapsedTime(ctx context.Context, action string, startTime time.Time, err error, optionalFields ...map[string]interface{}) {
	zapFields := logWithElapsedTime(ctx, startTime, optionalFields...)
	zapFields = append(zapFields, zap.Any("err", err))
	globalLogger.Error(action, zapFields...)
}
