/*
 * Copyright 2025 coze-dev Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package storage 提供对象存储接口
//
// 本包定义对象存储服务的接口，用于文件的存取：
// - 支持文件上传/下载
// - 支持文件删除
// - 支持生成预签名 URL
// - 支持文件列表查询
//
// 实现层在 impl/ 目录下，支持多种后端：
// - MinIO（本地开发）
// - S3（AWS 兼容）
// - TOS（火山引擎对象存储）
package storage

import (
	"context"
	"errors"
	"io"
	"time"
)

// ErrObjectNotFound 对象不存在错误
var (
	ErrObjectNotFound = errors.New("object not found")
)

// Storage 对象存储接口
//
//go:generate  mockgen -destination ../../internal/mock/infra/storage/storage_mock.go -package mock -source storage.go Factory
type Storage interface {
	// PutObject puts the object with the specified key.
	PutObject(ctx context.Context, objectKey string, content []byte, opts ...PutOptFn) error
	// PutObjectWithReader puts the object with the specified key.
	PutObjectWithReader(ctx context.Context, objectKey string, content io.Reader, opts ...PutOptFn) error
	// GetObject returns the object with the specified key.
	GetObject(ctx context.Context, objectKey string) ([]byte, error)
	// DeleteObject deletes the object with the specified key.
	DeleteObject(ctx context.Context, objectKey string) error
	// GetObjectUrl returns a presigned URL for the object.
	// The URL is valid for the specified duration.
	GetObjectUrl(ctx context.Context, objectKey string, opts ...GetOptFn) (string, error)
	// HeadObject returns the object metadata with the specified key.
	HeadObject(ctx context.Context, objectKey string, opts ...GetOptFn) (*FileInfo, error)
	// ListAllObjects returns all objects with the specified prefix.
	// It may return a large number of objects, consider using ListObjectsPaginated for better performance.
	ListAllObjects(ctx context.Context, prefix string, opts ...GetOptFn) ([]*FileInfo, error)
	// ListObjectsPaginated returns objects with pagination support.
	// Use this method when dealing with large number of objects.
	ListObjectsPaginated(ctx context.Context, input *ListObjectsPaginatedInput, opts ...GetOptFn) (*ListObjectsPaginatedOutput, error)
}

// SecurityToken 安全令牌，用于临时授权访问
type SecurityToken struct {
	AccessKeyID     string `thrift:"access_key_id,1" frugal:"1,default,string" json:"access_key_id"`
	SecretAccessKey string `thrift:"secret_access_key,2" frugal:"2,default,string" json:"secret_access_key"`
	SessionToken    string `thrift:"session_token,3" frugal:"3,default,string" json:"session_token"`
	ExpiredTime     string `thrift:"expired_time,4" frugal:"4,default,string" json:"expired_time"`
	CurrentTime     string `thrift:"current_time,5" frugal:"5,default,string" json:"current_time"`
}

// ListObjectsPaginatedInput 分页列表请求参数
type ListObjectsPaginatedInput struct {
	Prefix   string
	PageSize int
	Cursor   string
}

// ListObjectsPaginatedOutput 分页列表响应
type ListObjectsPaginatedOutput struct {
	Files  []*FileInfo
	Cursor string
	// false: All results have been returned
	// true: There are more results to return
	IsTruncated bool
}

// FileInfo 文件信息
type FileInfo struct {
	Key          string            `json:"key"`
	LastModified time.Time         `json:"last_modified"`
	ETag         string            `json:"etag"`
	Size         int64             `json:"size"`
	URL          string            `json:"url"`
	Tagging      map[string]string `json:"tagging"`
}
