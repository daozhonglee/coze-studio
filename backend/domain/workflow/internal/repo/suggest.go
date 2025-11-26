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

package repo

import (
	"context"
	"regexp"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	einoCompose "github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	"github.com/coze-dev/coze-studio/backend/bizpkg/llm/modelbuilder"
	"github.com/coze-dev/coze-studio/backend/domain/workflow"
	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity/vo"
	"github.com/coze-dev/coze-studio/backend/pkg/logs"
	"github.com/coze-dev/coze-studio/backend/pkg/sonic"
)

// SUGGESTION_PROMPT 后续问题建议生成的提示词模板
//
// 该提示词用于指导 LLM 根据用户问题和助手回答生成 3 个相关的后续问题。
// 支持通过 suggest_persona 模板变量注入角色设定。
const SUGGESTION_PROMPT = `
# Role
You are an AI assistant that quickly generates 3 relevant follow-up questions.

# Task
Analyze the user's question and the assistant's answer to suggest 3 unique, concise follow-up questions.

**IMPORTANT**: The assistant's answer can be very long. To be fast, focus only on the main ideas and topics in the answer. Do not analyze the full text in detail.

### Persona
{{ suggest_persona }}

## Output Format
- Return **only** a single JSON string array.
- Example: ["What is the history?", "How does it work?", "What are the benefits?"]
- The questions must be in the same language as the user's input.
`

// suggesterV3 后续问题建议生成器（V3 版本）
//
// 基于 Eino 框架实现，使用 LLM 生成后续问题建议。
type suggesterV3 struct {
	r einoCompose.Runnable[*vo.SuggestInfo, []string] // Eino 可执行链
}

// state 链执行过程中的状态
type state struct {
	userMessage *schema.Message // 用户输入消息
	answer      *schema.Message // 助手回答消息
}

// suggestRegexp 用于从 LLM 输出中提取 JSON 数组的正则表达式
var suggestRegexp = regexp.MustCompile(`\[(.*?)\]`)

// NewSuggester 创建后续问题建议生成器
//
// 使用 Eino 框架构建处理链：
// 1. 提取用户输入和助手回答
// 2. 应用提示词模板
// 3. 调用 LLM 生成建议
// 4. 解析 JSON 数组结果
//
// 参数：
//   - chatModel: 聊天模型实例
//
// 返回值：
//   - workflow.Suggester: 建议生成器接口实现
//   - error: 创建失败时返回错误
func NewSuggester(chatModel modelbuilder.BaseChatModel) (workflow.Suggester, error) {
	chain := einoCompose.NewChain[*vo.SuggestInfo, []string](einoCompose.WithGenLocalState(func(ctx context.Context) (s *state) {
		return &state{}
	}))
	r, err := chain.AppendLambda(einoCompose.InvokableLambda(func(ctx context.Context, input *vo.SuggestInfo) (output map[string]any, err error) {
		_ = compose.ProcessState(ctx, func(ctx context.Context, s *state) error {
			s.userMessage = input.UserInput
			s.answer = input.AnswerInput
			return nil
		})
		output = map[string]any{}
		if input.PersonaInput != nil {
			output["persona_input"] = *input.PersonaInput
		}
		return
	})).AppendChatTemplate(prompt.FromMessages(schema.Jinja2, schema.SystemMessage(SUGGESTION_PROMPT))).AppendChatModel(chatModel,
		compose.WithStatePreHandler(func(ctx context.Context, in []*schema.Message, state *state) ([]*schema.Message, error) {
			return append(in, []*schema.Message{state.userMessage, state.answer}...), nil
		})).AppendLambda(einoCompose.InvokableLambda(func(ctx context.Context, input *schema.Message) (output []string, err error) {
		content := suggestRegexp.FindString(input.Content)
		if len(content) == 0 {
			return
		}
		suggests := make([]string, 0)
		err = sonic.UnmarshalString(content, &suggests)
		if err != nil {
			logs.CtxErrorf(ctx, "Failed unmarshalling suggestions: %s", input.Content)

		}
		return suggests, nil
	})).Compile(context.Background())

	if err != nil {
		return nil, err
	}

	return &suggesterV3{r: r}, nil

}

// Suggest 生成后续问题建议
//
// 参数：
//   - ctx: 上下文
//   - info: 建议请求信息，包含用户输入和助手回答
//
// 返回值：
//   - []string: 生成的后续问题建议列表
//   - error: 生成失败时返回错误
func (s *suggesterV3) Suggest(ctx context.Context, info *vo.SuggestInfo) ([]string, error) {
	return s.r.Invoke(ctx, info)
}
