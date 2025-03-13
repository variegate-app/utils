package logger

import (
	"context"
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// fieldsKey - ключ контекста для хранения локальных полей в контексте.
type fieldsKey struct{}

// localFields - связывает строковые ключи с полями zap.Field.
// Используется для хранения локальных полей, которые могут быть добавлены к логгеру.
type localFields map[string]zap.Field

// Append добавляет новые поля которые могут быть добавлены к логгеру.
func (zf localFields) Append(fields ...zap.Field) localFields {
	zfCopy := make(localFields)
	for k, v := range zf {
		zfCopy[k] = v
	}

	for _, f := range fields {
		zfCopy[f.Key] = f
	}

	return zfCopy
}

type settings struct {
	config *zap.Config
	opts   []zap.Option
}

// defaultSettings создает настройки по умолчанию для zap.Logger.
// level - уровень логирования, который будет использоваться в настройках.
func defaultSettings(level zap.AtomicLevel) *settings {
	config := &zap.Config{
		Level:       level,
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: "json",
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:     "message",
			LevelKey:       "level",
			TimeKey:        "@timestamp",
			NameKey:        "logger",
			CallerKey:      "caller",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	return &settings{
		config: config,
		opts: []zap.Option{
			zap.AddCallerSkip(1),
		},
	}
}

type Instance struct {
	logger       *zap.Logger
	level        zap.AtomicLevel
	maskedFields map[string]struct{}
}

// New создает новый экземпляр логгера.
// level - уровень логирования, который будет использоваться в настройках.
// maskedFields - список полей, которые должны быть скрыты при логировании.
func New(level zapcore.Level, maskedFields ...string) (*Instance, error) {
	atomic := zap.NewAtomicLevelAt(level)
	settings := defaultSettings(atomic)

	l, err := settings.config.Build(settings.opts...)
	if err != nil {
		return nil, err
	}

	mf := make(map[string]struct{})
	for _, f := range maskedFields {
		mf[f] = struct{}{}
	}

	return &Instance{
		logger:       l,
		level:        atomic,
		maskedFields: mf,
	}, nil
}

// WithFields возвращает новый контекст с добавленными полями.
func (z *Instance) WithContextFields(ctx context.Context, fields ...zap.Field) context.Context {
	ctxFields, _ := ctx.Value(fieldsKey{}).(localFields)
	if ctxFields == nil {
		ctxFields = make(localFields)
	}

	merged := ctxFields.Append(fields...)
	return context.WithValue(ctx, fieldsKey{}, merged)
}

// maskField скрывает значение поля, если оно находится в списке скрытых полей.
func (z *Instance) maskField(f zap.Field) zap.Field {
	if _, ok := z.maskedFields[f.Key]; ok {
		return zap.String(f.Key, "******")
	}

	return f
}

// Sync синхронизирует буферы логгера и записывает все оставшиеся записи.
// Настоятельно рекомендуется вызывать его перед выходом из программы.
func (z *Instance) Sync() {
	_ = z.logger.Sync()
}

// withCtxFields добавляет поля контекста к переданным полям.
func (z *Instance) withCtxFields(ctx context.Context, fields ...zap.Field) []zap.Field {
	fs := make(localFields)

	ctxFields, _ := ctx.Value(fieldsKey{}).(localFields)
	if ctxFields != nil {
		fs = ctxFields
	}

	fs = fs.Append(fields...)

	var maskedFields []zap.Field
	for _, f := range fs {
		maskedFields = append(maskedFields, z.maskField(f))
	}

	return maskedFields
}

func (z *Instance) InfoCtx(ctx context.Context, msg string, fields ...zap.Field) {
	z.logger.Info(msg, z.withCtxFields(ctx, fields...)...)
}

func (z *Instance) DebugCtx(ctx context.Context, msg string, fields ...zap.Field) {
	z.logger.Debug(msg, z.withCtxFields(ctx, fields...)...)
}

func (z *Instance) WarnCtx(ctx context.Context, msg string, fields ...zap.Field) {
	z.logger.Warn(msg, z.withCtxFields(ctx, fields...)...)
}

func (z *Instance) ErrorCtx(ctx context.Context, msg string, fields ...zap.Field) {
	z.logger.Error(msg, z.withCtxFields(ctx, fields...)...)
}

func (z *Instance) FatalCtx(ctx context.Context, msg string, fields ...zap.Field) {
	z.logger.Fatal(msg, z.withCtxFields(ctx, fields...)...)
}

func (z *Instance) PanicCtx(ctx context.Context, msg string, fields ...zap.Field) {
	z.logger.Panic(msg, z.withCtxFields(ctx, fields...)...)
}

func (z *Instance) SetLevel(level zapcore.Level) {
	z.level.SetLevel(level)
}

// Std возвращает стандартный логгер, который может быть использован в коде, который не поддерживает zap.
func (z *Instance) Std() *log.Logger {
	return zap.NewStdLog(z.logger)
}
