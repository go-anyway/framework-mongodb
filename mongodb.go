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
	"fmt"
	"net/url"

	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBClient MongoDB 客户端封装
type MongoDBClient struct {
	Client   *mongo.Client
	Database *mongo.Database
	opts     *Options
}

// NewMongoDB 根据给定的选项创建一个新的 MongoDB 客户端实例
func NewMongoDB(opts *Options) (*MongoDBClient, error) {
	if opts == nil {
		return nil, fmt.Errorf("mongodb options cannot be nil")
	}
	return newMongoDB(opts)
}

// newMongoDB 内部函数，创建不带追踪的 MongoDB 客户端
func newMongoDB(opts *Options) (*MongoDBClient, error) {
	if opts == nil {
		return nil, fmt.Errorf("mongodb options cannot be nil")
	}

	// 构建连接 URI
	uri := buildURI(opts)

	// 创建客户端选项
	clientOptions := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(opts.MaxPoolSize).
		SetMinPoolSize(opts.MinPoolSize).
		SetConnectTimeout(opts.ConnectTimeout).
		SetSocketTimeout(opts.SocketTimeout)

	// 如果启用了追踪，添加 CommandMonitor
	if opts.EnableTrace {
		monitor := newTraceCommandMonitor(opts.EnableTrace)
		clientOptions.SetMonitor(&event.CommandMonitor{
			Started:   monitor.Started,
			Succeeded: monitor.Succeeded,
			Failed:    monitor.Failed,
		})
	}

	// 连接 MongoDB（使用推荐的 mongo.Connect 方式）
	ctx, cancel := context.WithTimeout(context.Background(), opts.ConnectTimeout)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongodb: %w", err)
	}

	// 测试连接
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping mongodb: %w", err)
	}

	// 获取数据库
	database := client.Database(opts.Database)

	return &MongoDBClient{
		Client:   client,
		Database: database,
		opts:     opts,
	}, nil
}

// buildURI 构建 MongoDB 连接 URI
func buildURI(opts *Options) string {
	// 构建基础 URI
	var uri string
	if opts.Username != "" && opts.Password != "" {
		// 有认证信息，对用户名和密码进行 URL 编码，防止特殊字符导致连接失败
		username := url.QueryEscape(opts.Username)
		password := url.QueryEscape(opts.Password)
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%d", username, password, opts.Host, opts.Port)
	} else {
		// 无认证信息
		uri = fmt.Sprintf("mongodb://%s:%d", opts.Host, opts.Port)
	}

	// 添加查询参数
	query := url.Values{}
	if opts.AuthSource != "" {
		query.Set("authSource", opts.AuthSource)
	}
	if len(query) > 0 {
		uri += "?" + query.Encode()
	}

	return uri
}

// Close 关闭 MongoDB 连接
func (c *MongoDBClient) Close(ctx context.Context) error {
	if c.Client == nil {
		return nil
	}
	return c.Client.Disconnect(ctx)
}

// Collection 获取集合
func (c *MongoDBClient) Collection(name string) *mongo.Collection {
	if c.Database == nil {
		return nil
	}
	return c.Database.Collection(name)
}

// Ping 测试连接
func (c *MongoDBClient) Ping(ctx context.Context) error {
	if c.Client == nil {
		return fmt.Errorf("mongodb client is not initialized")
	}
	return c.Client.Ping(ctx, nil)
}
