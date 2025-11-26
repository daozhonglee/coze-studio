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

// config.go 插件配置加载
//
// 本文件提供插件配置的初始化和加载功能：
//   - 插件产品元数据加载
//   - OAuth Schema 加载

package conf

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/bytedance/sonic"

	"github.com/coze-dev/coze-studio/backend/pkg/logs"
)

// InitConfig 初始化插件配置
// 从 resources/conf/plugin 目录加载配置文件
func InitConfig(ctx context.Context) (err error) {
	cwd, err := os.Getwd()
	if err != nil {
		logs.Warnf("[InitConfig] Failed to get current working directory: %v", err)
		cwd = os.Getenv("PWD")
	}

	basePath := path.Join(cwd, "resources", "conf", "plugin")

	err = loadPluginProductMeta(ctx, basePath)
	if err != nil {
		return err
	}

	err = loadOAuthSchema(ctx, basePath)
	if err != nil {
		return err
	}

	return nil
}

// oauthSchema OAuth 认证 Schema 配置
var oauthSchema string

// GetOAuthSchema 获取 OAuth Schema 配置
func GetOAuthSchema() string {
	return oauthSchema
}

// loadOAuthSchema 加载 OAuth Schema 配置文件
func loadOAuthSchema(ctx context.Context, basePath string) (err error) {
	filePath := path.Join(basePath, "common", "oauth_schema.json")
	file, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read file '%s' failed, err=%v", filePath, err)
	}

	if !isValidJSON(file) {
		return fmt.Errorf("invalid json, filePath=%s", filePath)
	}

	oauthSchema = string(file)

	return nil
}

// isValidJSON 检查数据是否为有效的 JSON
func isValidJSON(data []byte) bool {
	var js json.RawMessage
	return sonic.Unmarshal(data, &js) == nil
}
