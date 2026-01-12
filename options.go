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
	"fmt"
	"time"

	pkgConfig "github.com/go-anyway/framework-config"
)

// Config MongoDB 配置结构体（用于从配置文件创建）
type Config struct {
	Enabled        bool               `yaml:"enabled" env:"MONGODB_ENABLED" default:"true"`
	Host           string             `yaml:"host" env:"MONGODB_HOST" default:"localhost"`
	Port           int                `yaml:"port" env:"MONGODB_PORT" default:"27017"`
	Database       string             `yaml:"database" env:"MONGODB_DATABASE" required:"true"`
	Username       string             `yaml:"username" env:"MONGODB_USERNAME"`
	Password       string             `yaml:"password" env:"MONGODB_PASSWORD"`
	AuthSource     string             `yaml:"auth_source" env:"MONGODB_AUTH_SOURCE" default:"admin"`
	MaxPoolSize    int                `yaml:"max_pool_size" env:"MONGODB_MAX_POOL_SIZE" default:"100"`
	MinPoolSize    int                `yaml:"min_pool_size" env:"MONGODB_MIN_POOL_SIZE" default:"10"`
	ConnectTimeout pkgConfig.Duration `yaml:"connect_timeout" env:"MONGODB_CONNECT_TIMEOUT" default:"10s"`
	SocketTimeout  pkgConfig.Duration `yaml:"socket_timeout" env:"MONGODB_SOCKET_TIMEOUT" default:"30s"`
	EnableTrace    bool               `yaml:"enable_trace" env:"MONGODB_ENABLE_TRACE" default:"true"`
}

// Validate 验证 MongoDB 配置
func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("mongodb config cannot be nil")
	}
	if !c.Enabled {
		return nil // 如果未启用，不需要验证
	}
	if c.Database == "" {
		return fmt.Errorf("mongodb database is required")
	}
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("mongodb port must be between 1 and 65535, got %d", c.Port)
	}
	if c.MaxPoolSize < 1 {
		return fmt.Errorf("mongodb max_pool_size must be greater than 0, got %d", c.MaxPoolSize)
	}
	if c.MinPoolSize < 0 {
		return fmt.Errorf("mongodb min_pool_size must be non-negative, got %d", c.MinPoolSize)
	}
	if c.MaxPoolSize > 0 && c.MinPoolSize > 0 && c.MinPoolSize > c.MaxPoolSize {
		return fmt.Errorf("mongodb min_pool_size (%d) cannot be greater than max_pool_size (%d)", c.MinPoolSize, c.MaxPoolSize)
	}
	return nil
}

// ToOptions 转换为 Options
func (c *Config) ToOptions() (*Options, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	if !c.Enabled {
		return nil, fmt.Errorf("mongodb is not enabled")
	}

	connectTimeout := c.ConnectTimeout.Duration()
	if connectTimeout == 0 {
		connectTimeout = 10 * time.Second
	}
	socketTimeout := c.SocketTimeout.Duration()
	if socketTimeout == 0 {
		socketTimeout = 30 * time.Second
	}

	return &Options{
		Host:           c.Host,
		Port:           c.Port,
		Database:       c.Database,
		Username:       c.Username,
		Password:       c.Password,
		AuthSource:     c.AuthSource,
		MaxPoolSize:    uint64(c.MaxPoolSize),
		MinPoolSize:    uint64(c.MinPoolSize),
		ConnectTimeout: connectTimeout,
		SocketTimeout:  socketTimeout,
		EnableTrace:    c.EnableTrace,
	}, nil
}

// ConnectTimeoutDuration 返回 time.Duration 类型的 ConnectTimeout
func (c *Config) ConnectTimeoutDuration() time.Duration {
	return c.ConnectTimeout.Duration()
}

// SocketTimeoutDuration 返回 time.Duration 类型的 SocketTimeout
func (c *Config) SocketTimeoutDuration() time.Duration {
	return c.SocketTimeout.Duration()
}

// Options 结构体定义了 MongoDB 连接器的配置选项（内部使用）
type Options struct {
	Host           string
	Port           int
	Database       string
	Username       string
	Password       string
	AuthSource     string
	MaxPoolSize    uint64
	MinPoolSize    uint64
	ConnectTimeout time.Duration
	SocketTimeout  time.Duration
	EnableTrace    bool // 是否启用操作追踪，用于记录 MongoDB 操作执行时间
}
