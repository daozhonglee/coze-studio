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

package imagex

import (
	"context"
	"time"
)

//go:generate mockgen -destination ../../internal/mock/infra/imagex/imagex_mock.go --package imagex -source imagex.go

// ImageX 定义了图片上传和资源管理的接口
// 提供了获取上传凭证、上传文件、获取资源URL等功能
// 主要用于封装云存储服务（如AWS S3、阿里云OSS等）的上传和资源访问能力
type ImageX interface {
	// GetUploadAuth 获取上传认证令牌，使用默认过期时间
	GetUploadAuth(ctx context.Context, opt ...UploadAuthOpt) (*SecurityToken, error)

	// GetUploadAuthWithExpire 获取上传认证令牌，指定过期时间
	GetUploadAuthWithExpire(ctx context.Context, expire time.Duration, opt ...UploadAuthOpt) (*SecurityToken, error)

	// GetResourceURL 根据资源URI获取可访问的URL地址
	GetResourceURL(ctx context.Context, uri string, opts ...GetResourceOpt) (*ResourceURL, error)

	// Upload 直接上传文件数据到云存储
	Upload(ctx context.Context, data []byte, opts ...UploadAuthOpt) (*UploadResult, error)

	// GetServerID 获取服务端ID标识
	GetServerID() string

	// GetUploadHost 获取上传服务的主机地址
	GetUploadHost(ctx context.Context) string
}

// SecurityToken 安全令牌结构体，包含云存储访问所需的认证信息
type SecurityToken struct {
	AccessKeyID     string `thrift:"access_key_id,1" frugal:"1,default,string" json:"access_key_id"`         // 访问密钥ID
	SecretAccessKey string `thrift:"secret_access_key,2" frugal:"2,default,string" json:"secret_access_key"` // 秘密访问密钥
	SessionToken    string `thrift:"session_token,3" frugal:"3,default,string" json:"session_token"`         // 会话令牌
	ExpiredTime     string `thrift:"expired_time,4" frugal:"4,default,string" json:"expired_time"`           // 令牌过期时间
	CurrentTime     string `thrift:"current_time,5" frugal:"5,default,string" json:"current_time"`           // 当前时间
	HostScheme      string `thrift:"host_scheme,6" frugal:"6,default,string" json:"host_scheme"`             // 主机协议（http/https）
}

// ResourceURL 资源URL结构体，包含不同格式的资源访问地址
type ResourceURL struct {
	// REQUIRED; The resulting graph accesses the thin address, missing the bucket part compared to the default address.
	CompactURL string `json:"CompactURL"` // 紧凑URL格式，相对地址，不包含bucket信息
	// REQUIRED; Result graph access default address.
	URL string `json:"URL"` // 完整URL格式，包含完整的访问地址
}

// UploadResult 上传结果结构体，包含上传操作的完整返回信息
type UploadResult struct {
	Result    *Result   `json:"Results"`      // 上传结果详情
	RequestId string    `json:"RequestId"`    // 请求ID，用于跟踪和调试
	FileInfo  *FileInfo `json:"PluginResult"` // 文件信息详情
}

// Result 上传结果详情结构体
type Result struct {
	Uri       string `json:"Uri"`       // 上传文件的URI标识符
	UriStatus int    `json:"UriStatus"` // URI状态码，2000表示上传成功
}

// FileInfo 文件信息结构体，包含上传文件的详细信息和元数据
type FileInfo struct {
	Name        string `json:"FileName"`    // 文件名
	Uri         string `json:"ImageUri"`    // 文件URI
	ImageWidth  int    `json:"ImageWidth"`  // 图片宽度（像素）
	ImageHeight int    `json:"ImageHeight"` // 图片高度（像素）
	Md5         string `json:"ImageMd5"`    // 文件MD5哈希值
	ImageFormat string `json:"ImageFormat"` // 图片格式（如jpg、png等）
	ImageSize   int    `json:"ImageSize"`   // 文件大小（字节）
	FrameCnt    int    `json:"FrameCnt"`    // 帧数（用于动图）
	Duration    int    `json:"Duration"`    // 时长（用于视频或动图，单位毫秒）
}
