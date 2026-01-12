// Copyright 2025 zampo.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// @contact  zampo3380@gmail.com

package mongodb

import (
	"context"
	"errors"
	"strconv"
	"sync"

	"github.com/go-anyway/framework-log"
	pkgtrace "github.com/go-anyway/framework-trace"

	"go.mongodb.org/mongo-driver/event"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// traceCommandMonitor 实现 event.CommandMonitor 接口用于追踪命令（支持 OpenTelemetry）
type traceCommandMonitor struct {
	mu                 sync.RWMutex
	commandDatabaseMap map[int64]string     // RequestID -> DatabaseName 映射
	commandSpanMap     map[int64]trace.Span // RequestID -> Span 映射
	enableTrace        bool                 // 是否启用 OpenTelemetry 追踪
}

// newTraceCommandMonitor 创建追踪命令监视器
func newTraceCommandMonitor(enableTrace bool) *traceCommandMonitor {
	return &traceCommandMonitor{
		enableTrace: enableTrace,
	}
}

// Started 在命令开始时调用
func (m *traceCommandMonitor) Started(ctx context.Context, evt *event.CommandStartedEvent) {
	// 保存 RequestID 和 DatabaseName 的映射
	m.mu.Lock()
	if m.commandDatabaseMap == nil {
		m.commandDatabaseMap = make(map[int64]string)
	}
	m.commandDatabaseMap[evt.RequestID] = evt.DatabaseName

	// 如果启用了追踪，创建 OpenTelemetry span
	if m.enableTrace {
		if m.commandSpanMap == nil {
			m.commandSpanMap = make(map[int64]trace.Span)
		}
		spanName := "mongodb." + evt.CommandName
		_, span := pkgtrace.StartSpan(ctx, spanName,
			trace.WithAttributes(
				attribute.String("db.system", "mongodb"),
				attribute.String("db.name", evt.DatabaseName),
				attribute.String("db.operation", evt.CommandName),
			),
		)
		m.commandSpanMap[evt.RequestID] = span
		// 注意：这里无法更新原始 context，但 span 会通过 context 传播
	}
	m.mu.Unlock()
}

// Succeeded 在命令成功时调用
func (m *traceCommandMonitor) Succeeded(ctx context.Context, evt *event.CommandSucceededEvent) {
	duration := evt.Duration

	// 从映射中获取数据库名称和 span
	m.mu.Lock()
	dbName := m.commandDatabaseMap[evt.RequestID]
	delete(m.commandDatabaseMap, evt.RequestID)
	var span trace.Span
	if m.enableTrace && m.commandSpanMap != nil {
		span = m.commandSpanMap[evt.RequestID]
		delete(m.commandSpanMap, evt.RequestID)
	}
	m.mu.Unlock()

	// 如果启用了追踪，更新 span
	if m.enableTrace && span != nil {
		span.SetAttributes(
			attribute.Float64("db.duration_ms", float64(duration.Milliseconds())),
		)
		span.SetStatus(codes.Ok, "")
		span.End()
	}

	log.FromContext(ctx).Info(
		"MongoDB command success",
		zap.Float64("duration_ms", float64(duration.Milliseconds())),
		zap.String("command", evt.CommandName),
		zap.String("database", dbName),
		zap.String("request_id", strconv.FormatInt(evt.RequestID, 10)),
	)
}

// Failed 在命令失败时调用
func (m *traceCommandMonitor) Failed(ctx context.Context, evt *event.CommandFailedEvent) {
	duration := evt.Duration

	// 从映射中获取数据库名称和 span
	m.mu.Lock()
	dbName := m.commandDatabaseMap[evt.RequestID]
	delete(m.commandDatabaseMap, evt.RequestID)
	var span trace.Span
	if m.enableTrace && m.commandSpanMap != nil {
		span = m.commandSpanMap[evt.RequestID]
		delete(m.commandSpanMap, evt.RequestID)
	}
	m.mu.Unlock()

	// 如果启用了追踪，更新 span
	if m.enableTrace && span != nil {
		span.SetAttributes(
			attribute.Float64("db.duration_ms", float64(duration.Milliseconds())),
		)
		span.SetStatus(codes.Error, evt.Failure)
		span.RecordError(errors.New(evt.Failure))
		span.End()
	}

	log.FromContext(ctx).Error(
		"MongoDB command failed",
		zap.Float64("duration_ms", float64(duration.Milliseconds())),
		zap.String("command", evt.CommandName),
		zap.String("database", dbName),
		zap.String("request_id", strconv.FormatInt(evt.RequestID, 10)),
		zap.String("failure", evt.Failure),
	)
}
