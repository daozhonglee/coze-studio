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

// Package vo 定义了工作流画布的前端数据结构
//
// 这个包主要用于定义工作流的可视化编辑界面数据结构，包括：
// 1. Canvas - 工作流画布的整体定义
// 2. Node - 工作流中的节点定义（LLM、HTTP请求、数据库操作等）
// 3. Edge - 节点之间的连接关系
// 4. 各种节点类型的配置参数
//
// 在 Coze Studio 项目中，这些结构用于：
// - 前端工作流编辑器的渲染和交互
// - 工作流配置的序列化和反序列化
// - 工作流执行引擎的输入参数定义
// - 支持拖拽式工作流设计界面
package vo

import (
	"fmt"

	"github.com/coze-dev/coze-studio/backend/api/model/app/bot_common"
	"github.com/coze-dev/coze-studio/backend/api/model/workflow"
	"github.com/coze-dev/coze-studio/backend/pkg/i18n"
	"github.com/coze-dev/coze-studio/backend/pkg/lang/ternary"
)

// Canvas 定义了前端工作流画布的数据结构
// 这是工作流可视化编辑器的核心数据模型，包含了整个工作流的所有节点和连接关系
// 在项目中使用时，前端通过这个结构来渲染工作流图，后端通过它来解析和执行工作流
type Canvas struct {
	// Nodes 包含工作流中的所有节点，每个节点代表一个具体的执行单元（如LLM调用、HTTP请求等）
	Nodes []*Node `json:"nodes"`

	// Edges 定义节点之间的连接关系，表示数据流的方向
	Edges []*Edge `json:"edges"`

	// Versions 存储工作流画布的版本信息，用于兼容性处理
	Versions any `json:"versions"`
}

// Node 表示工作流画布中的一个节点
// 节点是工作流的基本执行单元，每个节点代表一个具体的操作（如LLM调用、HTTP请求、数据处理等）
// 在项目中，节点通过拖拽方式在前端画布上创建和配置
type Node struct {
	// ID 是工作流中节点的唯一标识符
	// 通常由前端生成，不需要在父子工作流之间保持唯一
	// 入口节点(Entry)和出口节点(Exit)的ID固定为：100001 和 900001
	ID string `json:"id"`

	// Type 是节点的类型，对应NodeMeta的ID字段值
	// 决定了节点的执行逻辑和配置项（如"llm"、"http"、"database"等）
	Type string `json:"type"`

	// Meta 存储前端使用的元数据，如节点在画布中的位置、大小等
	Meta any `json:"meta"`

	// Data 包含节点的实际配置信息，包括输入输出参数、异常处理等
	// 还包含不同节点类型的专属配置，如LLM的模型参数、HTTP请求的URL等
	Data *Data `json:"data"`

	// Blocks 存储复合节点的子节点
	// 仅在节点类型为复合类型时使用，如批量处理(Batch)和循环(Loop)节点
	Blocks []*Node `json:"blocks,omitempty"`

	// Edges 存储节点内部的连接关系
	// 严格对应画布上绘制的连接线，主要用于复合节点内部的数据流
	Edges []*Edge `json:"edges,omitempty"`

	// Version 是此节点类型schema的版本号
	// 用于处理不同版本的节点配置兼容性
	Version string `json:"version,omitempty"`

	// parent 指向父节点，仅在节点位于复合节点内部时设置
	parent *Node
}

func (n *Node) SetParent(parent *Node) {
	n.parent = parent
}

func (n *Node) Parent() *Node {
	return n.parent
}

// NodeMetaFE 定义前端显示用的节点元数据
// 这些信息用于在工作流编辑器中渲染节点的外观和基本信息
type NodeMetaFE struct {
	Title       string `json:"title,omitempty"`       // 节点标题，在画布上显示的名称
	Description string `json:"description,omitempty"` // 节点描述，鼠标悬停时显示的提示信息
	Icon        string `json:"icon,omitempty"`        // 节点图标的URL地址
	SubTitle    string `json:"subTitle,omitempty"`    // 副标题，显示在主标题下方的次级信息
	MainColor   string `json:"mainColor,omitempty"`   // 主色调，用于节点边框和图标的颜色
}

// Edge 定义节点之间的连接关系
// 表示工作流中数据流的方向，从源节点流向目标节点
type Edge struct {
	SourceNodeID string `json:"sourceNodeID"`           // 源节点ID，数据流出的节点
	TargetNodeID string `json:"targetNodeID"`           // 目标节点ID，数据流入的节点
	SourcePortID string `json:"sourcePortID,omitempty"` // 源端口ID，指定数据从哪个输出端口流出
	TargetPortID string `json:"targetPortID,omitempty"` // 目标端口ID，指定数据流入哪个输入端口
}

// Data 包含节点的实际配置数据
// 这是节点的核心配置，定义了节点的输入输出参数、异常处理等所有运行时需要的配置信息
type Data struct {
	// Meta 存储节点的元数据，仅前端使用
	Meta *NodeMetaFE `json:"nodeMeta,omitempty"`

	// Outputs 配置节点的输出字段及其类型
	// 可以是[]*Variable（大多数情况，仅字段和类型）或[]*Param（复合节点使用，需要引用子节点输出）
	Outputs []any `json:"outputs,omitempty"`

	// Inputs 配置节点的所有输入信息
	// 包括固定的输入字段和用户自定义的动态输入字段
	Inputs *Inputs `json:"inputs,omitempty"`

	// Size 配置节点在前端显示的大小
	// 仅用于NodeTypeComment类型的节点
	Size any `json:"size,omitempty"`
}

// Inputs 包含节点的所有输入配置信息
// 这个结构体聚合了各种类型的节点配置，每个指针字段对应一种特定的节点类型
type Inputs struct {
	// InputParameters 是用户为此节点定义的输入字段
	InputParameters []*Param `json:"inputParameters"`

	// ChatHistorySetting 配置聊天流模式下此节点的聊天历史设置
	ChatHistorySetting *ChatHistorySetting `json:"chatHistorySetting,omitempty"`

	// SettingOnError 配置节点的通用错误处理策略
	// 注意：需要在前端节点表单中先启用
	SettingOnError *SettingOnError `json:"settingOnError,omitempty"`

	// NodeBatchInfo 配置节点的批量处理模式
	// 注意：不要与NodeTypeBatch类型混淆，这是单个节点的批量配置
	NodeBatchInfo *NodeBatch `json:"batch,omitempty"`

	// LLMParam 可能是LLMParam、IntentDetectorLLMParam或SimpleLLMParam之一
	// 在大多数需要ChatModel功能的节点之间共享
	LLMParam any `json:"llmParam,omitempty"`

	// 以下是指向各种专用节点类型配置的指针
	*OutputEmitter      // NodeTypeEmitter和Answer模式下NodeTypeExit的专用配置
	*Exit               // NodeTypeExit的专用配置
	*LLM                // NodeTypeLLM的专用配置
	*Loop               // NodeTypeLoop的专用配置
	*Selector           // NodeTypeSelector的专用配置
	*TextProcessor      // NodeTypeTextProcessor的专用配置
	*SubWorkflow        // NodeTypeSubWorkflow的专用配置
	*IntentDetector     // NodeTypeIntentDetector的专用配置
	*DatabaseNode       // 各种数据库节点类型的专用配置
	*HttpRequestNode    // NodeTypeHTTPRequester的专用配置
	*Knowledge          // 各种知识库节点类型的专用配置
	*CodeRunner         // NodeTypeCodeRunner的专用配置
	*PluginAPIParam     // NodeTypePlugin的专用配置
	*VariableAggregator // NodeTypeVariableAggregator的专用配置
	*VariableAssigner   // NodeTypeVariableAssigner的专用配置
	*QA                 // NodeTypeQuestionAnswer的专用配置
	*Batch              // NodeTypeBatch的专用配置
	*Comment            // NodeTypeComment的专用配置
	*InputReceiver      // NodeTypeInputReceiver的专用配置
}

// OutputEmitter 配置输出发射器节点（用于Answer模式下的内容输出）
type OutputEmitter struct {
	Content         *BlockInput `json:"content"`                   // 要输出的内容
	StreamingOutput bool        `json:"streamingOutput,omitempty"` // 是否启用流式输出
}

// Exit 配置工作流出口节点
type Exit struct {
	TerminatePlan *TerminatePlan `json:"terminatePlan,omitempty"` // 工作流终止计划（返回变量或使用回答内容）
}

// LLM 配置大语言模型节点
type LLM struct {
	FCParam *FCParam `json:"fcParam,omitempty"` // 函数调用参数配置
}

type Loop struct {
	LoopType           LoopType    `json:"loopType,omitempty"`
	LoopCount          *BlockInput `json:"loopCount,omitempty"`
	VariableParameters []*Param    `json:"variableParameters,omitempty"`
}

type Selector struct {
	Branches []*struct {
		Condition struct {
			Logic      LogicType    `json:"logic"`
			Conditions []*Condition `json:"conditions"`
		} `json:"condition"`
	} `json:"branches,omitempty"`
}

type Comment struct {
	SchemaType string `json:"schemaType,omitempty"`
	Note       any    `json:"note,omitempty"`
}

type TextProcessor struct {
	Method       TextProcessingMethod `json:"method,omitempty"`
	ConcatParams []*Param             `json:"concatParams,omitempty"`
	SplitParams  []*Param             `json:"splitParams,omitempty"`
}

type VariableAssigner struct {
	VariableTypeMap map[string]any `json:"variableTypeMap,omitempty"`
}

type LLMParam = []*Param
type IntentDetectorLLMParam = map[string]any
type SimpleLLMParam struct {
	GenerationDiversity string         `json:"generationDiversity"`
	MaxTokens           int            `json:"maxTokens"`
	ModelName           string         `json:"modelName"`
	ModelType           int64          `json:"modelType"`
	ResponseFormat      ResponseFormat `json:"responseFormat"`
	SystemPrompt        string         `json:"systemPrompt"`
	Temperature         float64        `json:"temperature"`
	TopP                float64        `json:"topP"`
}

type QA struct {
	AnswerType    QAAnswerType `json:"answer_type"`
	Limit         int          `json:"limit,omitempty"`
	ExtractOutput bool         `json:"extra_output,omitempty"`
	OptionType    QAOptionType `json:"option_type,omitempty"`
	Options       []struct {
		Name string `json:"name"`
	} `json:"options,omitempty"`
	Question      string      `json:"question,omitempty"`
	DynamicOption *BlockInput `json:"dynamic_option,omitempty"`
}

type QAAnswerType string

const (
	QAAnswerTypeOption QAAnswerType = "option"
	QAAnswerTypeText   QAAnswerType = "text"
)

type QAOptionType string

const (
	QAOptionTypeStatic  QAOptionType = "static"
	QAOptionTypeDynamic QAOptionType = "dynamic"
)

type RequestParameter struct {
	Name string
}

type FCParam struct {
	WorkflowFCParam *struct {
		WorkflowList []struct {
			WorkflowID      string `json:"workflow_id"`
			WorkflowVersion string `json:"workflow_version"`
			PluginID        string `json:"plugin_id"`
			PluginVersion   string `json:"plugin_version"`
			IsDraft         bool   `json:"is_draft"`
			FCSetting       *struct {
				RequestParameters  []*workflow.APIParameter `json:"request_params"`
				ResponseParameters []*workflow.APIParameter `json:"response_params"`
			} `json:"fc_setting,omitempty"`
		} `json:"workflowList,omitempty"`
	} `json:"workflowFCParam,omitempty"`
	PluginFCParam *struct {
		PluginList []struct {
			PluginID      string `json:"plugin_id"`
			ApiId         string `json:"api_id"`
			ApiName       string `json:"api_name"`
			PluginVersion string `json:"plugin_version"`
			IsDraft       bool   `json:"is_draft"`

			PluginFrom *bot_common.PluginFrom `json:"plugin_from"`
			FCSetting  *struct {
				RequestParameters  []*workflow.APIParameter `json:"request_params"`
				ResponseParameters []*workflow.APIParameter `json:"response_params"`
			} `json:"fc_setting,omitempty"`
		} `json:"pluginList,omitempty"`
	} `json:"pluginFCParam,omitempty"`

	KnowledgeFCParam *struct {
		GlobalSetting *struct {
			SearchMode                   int64   `json:"search_mode"`
			TopK                         int64   `json:"top_k"`
			MinScore                     float64 `json:"min_score"`
			UseNL2SQL                    bool    `json:"use_nl2_sql"`
			UseRewrite                   bool    `json:"use_rewrite"`
			UseRerank                    bool    `json:"use_rerank"`
			NoRecallReplyCustomizePrompt string  `json:"no_recall_reply_customize_prompt"`
			NoRecallReplyMode            int64   `json:"no_recall_reply_mode"`
		} `json:"global_setting,omitempty"`
		KnowledgeList []*struct {
			ID string `json:"id"`
		} `json:"knowledgeList,omitempty"`
	} `json:"knowledgeFCParam,omitempty"`
}

type Batch struct {
	BatchSize      *BlockInput `json:"batchSize,omitempty"`
	ConcurrentSize *BlockInput `json:"concurrentSize,omitempty"`
}

type NodeBatch struct {
	BatchEnable    bool     `json:"batchEnable"`
	BatchSize      int64    `json:"batchSize"`
	ConcurrentSize int64    `json:"concurrentSize"`
	InputLists     []*Param `json:"inputLists,omitempty"`
}

type IntentDetectorLLMConfig struct {
	ModelName      string     `json:"modelName"`
	ModelType      int        `json:"modelType"`
	Temperature    *float64   `json:"temperature"`
	TopP           *float64   `json:"topP"`
	MaxTokens      int        `json:"maxTokens"`
	ResponseFormat int64      `json:"responseFormat"`
	SystemPrompt   BlockInput `json:"systemPrompt"`
}

type VariableAggregator struct {
	MergeGroups []*Param `json:"mergeGroups,omitempty"`
}

type PluginAPIParam struct {
	APIParams  []*Param               `json:"apiParam"`
	PluginFrom *bot_common.PluginFrom `json:"pluginFrom"`
}

type CodeRunner struct {
	Code     string `json:"code"`
	Language int64  `json:"language"`
}

type Knowledge struct {
	DatasetParam  []*Param      `json:"datasetParam,omitempty"`
	StrategyParam StrategyParam `json:"strategyParam,omitempty"`
}

type StrategyParam struct {
	ParsingStrategy struct {
		ParsingType     string `json:"parsingType,omitempty"`
		ImageExtraction bool   `json:"imageExtraction"`
		TableExtraction bool   `json:"tableExtraction"`
		ImageOcr        bool   `json:"imageOcr"`
	} `json:"parsingStrategy,omitempty"`
	ChunkStrategy struct {
		ChunkType     string  `json:"chunkType,omitempty"`
		SeparatorType string  `json:"separatorType,omitempty"`
		Separator     string  `json:"separator,omitempty"`
		MaxToken      int64   `json:"maxToken,omitempty"`
		Overlap       float64 `json:"overlap,omitempty"`
	} `json:"chunkStrategy,omitempty"`
	IndexStrategy any `json:"indexStrategy"`
}

// HttpRequestNode 配置HTTP请求节点
// 支持GET、POST等HTTP方法，以及复杂的请求配置（认证、超时、重试等）
type HttpRequestNode struct {
	APIInfo APIInfo             `json:"apiInfo,omitempty"` // API基本信息（方法、URL）
	Body    Body                `json:"body,omitempty"`    // 请求体配置
	Headers []*Param            `json:"headers"`           // 请求头参数
	Params  []*Param            `json:"params"`            // URL查询参数
	Auth    *Auth               `json:"auth"`              // 认证配置
	Setting *HttpRequestSetting `json:"setting"`           // 请求设置（超时、重试等）
}

type APIInfo struct {
	Method string `json:"method"`
	URL    string `json:"url"`
}
type Body struct {
	BodyType string    `json:"bodyType"`
	BodyData *BodyData `json:"bodyData"`
}
type BodyData struct {
	Json     string `json:"json,omitempty"`
	FormData *struct {
		Data []*Param `json:"data"`
	} `json:"formData,omitempty"`
	FormURLEncoded []*Param `json:"formURLEncoded,omitempty"`
	RawText        string   `json:"rawText,omitempty"`
	Binary         struct {
		FileURL *BlockInput `json:"fileURL"`
	} `json:"binary"`
}

type Auth struct {
	AuthType string `json:"authType"`
	AuthData struct {
		CustomData struct {
			AddTo string   `json:"addTo"`
			Data  []*Param `json:"data,omitempty"`
		} `json:"customData"`
		BearerTokenData []*Param `json:"bearerTokenData,omitempty"`
	} `json:"authData"`

	AuthOpen bool `json:"authOpen"`
}

type HttpRequestSetting struct {
	Timeout    int64 `json:"timeout"`
	RetryTimes int64 `json:"retryTimes"`
}

type DatabaseNode struct {
	DatabaseInfoList []*DatabaseInfo `json:"databaseInfoList,omitempty"`
	SQL              string          `json:"sql,omitempty"`
	SelectParam      *SelectParam    `json:"selectParam,omitempty"`

	InsertParam *InsertParam `json:"insertParam,omitempty"`

	DeleteParam *DeleteParam `json:"deleteParam,omitempty"`

	UpdateParam *UpdateParam `json:"updateParam,omitempty"`
}

type DatabaseLogicType string

const (
	DatabaseLogicAnd DatabaseLogicType = "AND"
	DatabaseLogicOr  DatabaseLogicType = "OR"
)

type DBCondition struct {
	ConditionList [][]*Param        `json:"conditionList,omitempty"`
	Logic         DatabaseLogicType `json:"logic"`
}

type UpdateParam struct {
	Condition DBCondition `json:"condition"`
	FieldInfo [][]*Param  `json:"fieldInfo"`
}

type DeleteParam struct {
	Condition DBCondition `json:"condition"`
}

type InsertParam struct {
	FieldInfo [][]*Param `json:"fieldInfo"`
}

type SelectParam struct {
	Condition   *DBCondition `json:"condition,omitempty"` // may be nil
	OrderByList []struct {
		FieldID int64 `json:"fieldID"`
		IsAsc   bool  `json:"isAsc"`
	} `json:"orderByList,omitempty"`
	Limit     int64 `json:"limit"`
	FieldList []struct {
		FieldID    int64 `json:"fieldID"`
		IsDistinct bool  `json:"isDistinct"`
	} `json:"fieldList,omitempty"`
}

type DatabaseInfo struct {
	DatabaseInfoID string `json:"databaseInfoID"`
}

type IntentDetector struct {
	Intents []*Intent `json:"intents,omitempty"`
	Mode    string    `json:"mode,omitempty"`
}
type ChatHistorySetting struct {
	EnableChatHistory bool  `json:"enableChatHistory,omitempty"`
	ChatHistoryRound  int64 `json:"chatHistoryRound,omitempty"`
}

type Intent struct {
	Name string `json:"name"`
}

// Param is a node's field with type and source info.
type Param struct {
	// Name is the field's name.
	Name string `json:"name,omitempty"`

	// Input is the configurations for normal, singular field.
	Input *BlockInput `json:"input,omitempty"`

	// Left is the configurations for the left half of an expression,
	// such as an assignment in NodeTypeVariableAssigner.
	Left *BlockInput `json:"left,omitempty"`

	// Right is the configuration for the right half of an expression.
	Right *BlockInput `json:"right,omitempty"`

	// Variables are configurations for a group of fields.
	// Only used in NodeTypeVariableAggregator.
	Variables []*BlockInput `json:"variables,omitempty"`
}

// Variable is the configuration of a node's field, either input or output.
type Variable struct {
	// Name is the field's name as defined on canvas.
	Name string `json:"name"`

	// Type is the field's data type, such as string, integer, number, object, array, etc.
	Type VariableType `json:"type"`

	// Required is set to true if you checked the 'required box' on this field
	Required bool `json:"required,omitempty"`

	// AssistType is the 'secondary' type of string fields, such as different types of file and image, or time.
	AssistType AssistType `json:"assistType,omitempty"`

	// Schema contains detailed info for sub-fields of an object field, or element type of an array.
	Schema any `json:"schema,omitempty"` // either []*Variable (for object) or *Variable (for list)

	// Description describes the field's intended use. Used on Entry node. Useful for workflow tools.
	Description string `json:"description,omitempty"`

	// ReadOnly indicates a field is not to be set by Node's business logic.
	// e.g. the ErrorBody field when exception strategy is configured.
	ReadOnly bool `json:"readOnly,omitempty"`

	// DefaultValue configures the 'default value' if this field is missing in input.
	// Effective only in Entry node.
	DefaultValue any `json:"defaultValue,omitempty"`
}

type BlockInput struct {
	Type       VariableType     `json:"type,omitempty" yaml:"Type,omitempty"`
	AssistType AssistType       `json:"assistType,omitempty" yaml:"AssistType,omitempty"`
	Schema     any              `json:"schema,omitempty" yaml:"Schema,omitempty"` // either *BlockInput(or *Variable) for list or []*Variable (for object)
	Value      *BlockInputValue `json:"value,omitempty" yaml:"Value,omitempty"`
}

type BlockInputValue struct {
	Type    BlockInputValueType `json:"type"`
	Content any                 `json:"content,omitempty"` // either string for text such as template, or BlockInputReference
	RawMeta any                 `json:"rawMeta,omitempty"`
}

type BlockInputReference struct {
	BlockID string        `json:"blockID"`
	Name    string        `json:"name,omitempty"`
	Path    []string      `json:"path,omitempty"`
	Source  RefSourceType `json:"source"`
}

type Condition struct {
	Operator OperatorType `json:"operator"`
	Left     *Param       `json:"left"`
	Right    *Param       `json:"right,omitempty"`
}

type SubWorkflow struct {
	WorkflowID      string `json:"workflowId,omitempty"`
	WorkflowVersion string `json:"workflowVersion,omitempty"`
	TerminationType int    `json:"type,omitempty"`
	SpaceID         string `json:"spaceId,omitempty"`
}

// VariableType 定义变量的数据类型
// 用于指定工作流中变量的类型，支持基本类型和复合类型
type VariableType string

const (
	VariableTypeString  VariableType = "string"  // 字符串类型
	VariableTypeInteger VariableType = "integer" // 整数类型
	VariableTypeFloat   VariableType = "float"   // 浮点数类型
	VariableTypeBoolean VariableType = "boolean" // 布尔类型
	VariableTypeObject  VariableType = "object"  // 对象类型（JSON对象）
	VariableTypeList    VariableType = "list"    // 列表类型（数组）
)

type AssistType = int64

const (
	AssistTypeNotSet  AssistType = 0
	AssistTypeDefault AssistType = 1
	AssistTypeImage   AssistType = 2
	AssistTypeDoc     AssistType = 3
	AssistTypeCode    AssistType = 4
	AssistTypePPT     AssistType = 5
	AssistTypeTXT     AssistType = 6
	AssistTypeExcel   AssistType = 7
	AssistTypeAudio   AssistType = 8
	AssistTypeZip     AssistType = 9
	AssistTypeVideo   AssistType = 10
	AssistTypeSvg     AssistType = 11
	AssistTypeVoice   AssistType = 12

	AssistTypeTime AssistType = 10000
)

type BlockInputValueType string

const (
	BlockInputValueTypeLiteral   BlockInputValueType = "literal"
	BlockInputValueTypeRef       BlockInputValueType = "ref"
	BlockInputValueTypeObjectRef BlockInputValueType = "object_ref"
)

type RefSourceType string

const (
	RefSourceTypeBlockOutput  RefSourceType = "block-output" // Represents an implicitly declared variable that references the output of a block
	RefSourceTypeGlobalApp    RefSourceType = "global_variable_app"
	RefSourceTypeGlobalSystem RefSourceType = "global_variable_system"
	RefSourceTypeGlobalUser   RefSourceType = "global_variable_user"
)

type TerminatePlan string

const (
	ReturnVariables  TerminatePlan = "returnVariables"
	UseAnswerContent TerminatePlan = "useAnswerContent"
)

type ErrorProcessType int

const (
	ErrorProcessTypeThrow             ErrorProcessType = 1 // throws the error as usual
	ErrorProcessTypeReturnDefaultData ErrorProcessType = 2 // return DataOnErr configured in SettingOnError
	ErrorProcessTypeExceptionBranch   ErrorProcessType = 3 // executes the exception branch on error
)

// SettingOnError contains common error handling strategy.
type SettingOnError struct {
	// DataOnErr defines the JSON result to be returned on error.
	DataOnErr string `json:"dataOnErr,omitempty"`
	// Switch defines whether ANY error handling strategy is active.
	// If set to false, it's equivalent to set ProcessType = ErrorProcessTypeThrow
	Switch bool `json:"switch,omitempty"`
	// ProcessType determines the error handling strategy for this node.
	ProcessType *ErrorProcessType `json:"processType,omitempty"`
	// RetryTimes determines how many times to retry. 0 means no retry.
	// If positive, any retries will be executed immediately after error.
	RetryTimes int64 `json:"retryTimes,omitempty"`
	// TimeoutMs sets the timeout duration in millisecond.
	// If any retry happens, ALL retry attempts accumulates to the same timeout threshold.
	TimeoutMs int64 `json:"timeoutMs,omitempty"`
	// Ext sets any extra settings specific to NodeType
	Ext *struct {
		// BackupLLMParam is only for LLM Node, marshaled from SimpleLLMParam.
		// If retry happens, the backup LLM will be used instead of the main LLM.
		BackupLLMParam string `json:"backupLLMParam,omitempty"`
	} `json:"ext,omitempty"`
}

type LogicType int

const (
	_ LogicType = iota
	OR
	AND
)

// OperatorType 定义条件判断的操作符类型
// 用于选择器节点中的条件判断逻辑
type OperatorType int

const (
	_                      OperatorType = iota
	Equal                               // 等于
	NotEqual                            // 不等于
	LengthGreaterThan                   // 长度大于
	LengthGreaterThanEqual              // 长度大于等于
	LengthLessThan                      // 长度小于
	LengthLessThanEqual                 // 长度小于等于
	Contain                             // 包含
	NotContain                          // 不包含
	Empty                               // 为空
	NotEmpty                            // 不为空
	True                                // 为真
	False                               // 为假
	GreaterThan                         // 大于
	GreaterThanEqual                    // 大于等于
	LessThan                            // 小于
	LessThanEqual                       // 小于等于
)

type TextProcessingMethod string

const (
	Concat TextProcessingMethod = "concat"
	Split  TextProcessingMethod = "split"
)

type LoopType string

const (
	LoopTypeArray    LoopType = "array"
	LoopTypeCount    LoopType = "count"
	LoopTypeInfinite LoopType = "infinite"
)

type InputReceiver struct {
	OutputSchema string `json:"outputSchema,omitempty"`
}

// GenerateNodeIDForBatchMode 为批量模式生成内部节点ID
// 在批量处理节点中，内部节点会添加"_inner"后缀来区分
func GenerateNodeIDForBatchMode(key string) string {
	return key + "_inner"
}

// IsGeneratedNodeForBatchMode 检查节点是否为批量模式生成的内部节点
func IsGeneratedNodeForBatchMode(key string, parentKey string) bool {
	return key == GenerateNodeIDForBatchMode(parentKey)
}

const defaultZhCNInitCanvasJsonSchema = `{
 "nodes": [
  {
   "id": "100001",
   "type": "1",
   "meta": {
    "position": {
     "x": 0,
     "y": 0
    }
   },
   "data": {
    "nodeMeta": {
     "description": "工作流的起始节点，用于设定启动工作流需要的信息",
     "icon": "https://lf3-static.bytednsdoc.com/obj/eden-cn/dvsmryvd_avi_dvsm/ljhwZthlaukjlkulzlp/icon/icon-Start.png",
     "subTitle": "",
     "title": "开始"
    },
    "outputs": [
     {
      "type": "string",
      "name": "input",
      "required": false
     }
    ],
    "trigger_parameters": [
     {
      "type": "string",
      "name": "input",
      "required": false
     }
    ]
   }
  },
  {
   "id": "900001",
   "type": "2",
   "meta": {
    "position": {
     "x": 1000,
     "y": 0
    }
   },
   "data": {
    "nodeMeta": {
     "description": "工作流的最终节点，用于返回工作流运行后的结果信息",
     "icon": "https://lf3-static.bytednsdoc.com/obj/eden-cn/dvsmryvd_avi_dvsm/ljhwZthlaukjlkulzlp/icon/icon-End.png",
     "subTitle": "",
     "title": "结束"
    },
    "inputs": {
     "terminatePlan": "returnVariables",
     "inputParameters": [
      {
       "name": "output",
       "input": {
        "type": "string",
        "value": {
         "type": "ref",
         "content": {
          "source": "block-output",
          "blockID": "",
          "name": ""
         }
        }
       }
      }
     ]
    }
   }
  }
 ],
 "edges": [],
 "versions": {
  "loop": "v2"
 }
}`

const defaultEnUSInitCanvasJsonSchema = `{
 "nodes": [
  {
   "id": "100001",
   "type": "1",
   "meta": {
    "position": {
     "x": 0,
     "y": 0
    }
   },
   "data": {
    "nodeMeta": {
     "description": "The starting node of the workflow, used to set the information needed to initiate the workflow.",
     "icon": "https://lf3-static.bytednsdoc.com/obj/eden-cn/dvsmryvd_avi_dvsm/ljhwZthlaukjlkulzlp/icon/icon-Start.png",
     "subTitle": "",
     "title": "Start"
    },
    "outputs": [
     {
      "type": "string",
      "name": "input",
      "required": false
     }
    ],
    "trigger_parameters": [
     {
      "type": "string",
      "name": "input",
      "required": false
     }
    ]
   }
  },
  {
   "id": "900001",
   "type": "2",
   "meta": {
    "position": {
     "x": 1000,
     "y": 0
    }
   },
   "data": {
    "nodeMeta": {
     "description": "The final node of the workflow, used to return the result information after the workflow runs.",
     "icon": "https://lf3-static.bytednsdoc.com/obj/eden-cn/dvsmryvd_avi_dvsm/ljhwZthlaukjlkulzlp/icon/icon-End.png",
     "subTitle": "",
     "title": "End"
    },
    "inputs": {
     "terminatePlan": "returnVariables",
     "inputParameters": [
      {
       "name": "output",
       "input": {
        "type": "string",
        "value": {
         "type": "ref",
         "content": {
          "source": "block-output",
          "blockID": "",
          "name": ""
         }
        }
       }
      }
     ]
    }
   }
  }
 ],
 "edges": [],
 "versions": {
  "loop": "v2"
 }
}`

const defaultZhCNInitCanvasJsonSchemaChat = `{
	"nodes": [{
		"id": "100001",
		"type": "1",
		"meta": {
			"position": {
				"x": 0,
				"y": 0
			}
		},
		"data": {
			"outputs": [{
				"type": "string",
				"name": "USER_INPUT",
				"required": true
			}, {
				"type": "string",
				"name": "CONVERSATION_NAME",
				"required": false,
				"description": "本次请求绑定的会话，会自动写入消息、会从该会话读对话历史。",
				"defaultValue": "%s"
			}],
			"nodeMeta": {
				"title": "开始",
				"icon": "https://lf3-static.bytednsdoc.com/obj/eden-cn/dvsmryvd_avi_dvsm/ljhwZthlaukjlkulzlp/icon/icon-Start.png",
				"description": "工作流的起始节点，用于设定启动工作流需要的信息",
				"subTitle": ""
			}
		}
	}, {
		"id": "900001",
		"type": "2",
		"meta": {
			"position": {
				"x": 1000,
				"y": 0
			}
		},
		"data": {
			"nodeMeta": {
				"title": "结束",
				"icon": "https://lf3-static.bytednsdoc.com/obj/eden-cn/dvsmryvd_avi_dvsm/ljhwZthlaukjlkulzlp/icon/icon-End.png",
				"description": "工作流的最终节点，用于返回工作流运行后的结果信息",
				"subTitle": ""
			},
			"inputs": {
				"terminatePlan": "useAnswerContent",
				"streamingOutput": true,
				"inputParameters": [{
					"name": "output",
					"input": {
						"type": "string",
						"value": {
							"type": "ref"
						}
					}
				}]
			}
		}
	}]
}`
const defaultEnUSInitCanvasJsonSchemaChat = `{
	"nodes": [{
		"id": "100001",
		"type": "1",
		"meta": {
			"position": {
				"x": 0,
				"y": 0
			}
		},
		"data": {
			"outputs": [{
				"type": "string",
				"name": "USER_INPUT",
				"required": true
			}, {
				"type": "string",
				"name": "CONVERSATION_NAME",
				"required": false,
				"description": "The conversation bound to this request will automatically write messages and read conversation history from that conversation.",
				"defaultValue": "%s"
			}],
			"nodeMeta": {
				"title": "Start",
				"icon": "https://lf3-static.bytednsdoc.com/obj/eden-cn/dvsmryvd_avi_dvsm/ljhwZthlaukjlkulzlp/icon/icon-Start.png",
				"description": "The starting node of the workflow, used to set the information needed to initiate the workflow.",
				"subTitle": ""
			}
		}
	}, {
		"id": "900001",
		"type": "2",
		"meta": {
			"position": {
				"x": 1000,
				"y": 0
			}
		},
		"data": {
			"nodeMeta": {
				"title": "End",
				"icon": "https://lf3-static.bytednsdoc.com/obj/eden-cn/dvsmryvd_avi_dvsm/ljhwZthlaukjlkulzlp/icon/icon-End.png",
				"description": "The final node of the workflow, used to return the result information after the workflow runs.",
				"subTitle": ""
			},
			"inputs": {
				"terminatePlan": "useAnswerContent",
				"streamingOutput": true,
				"inputParameters": [{
					"name": "output",
					"input": {
						"type": "string",
						"value": {
							"type": "ref"
						}
					}
				}]
			}
		}
	}]
}`

// GetDefaultInitCanvasJsonSchema 根据语言环境获取默认的工作流画布配置
// 返回包含开始和结束节点的完整工作流模板，用于新建工作流时的初始化
func GetDefaultInitCanvasJsonSchema(locale i18n.Locale) string {
	return ternary.IFElse(locale == i18n.LocaleEN, defaultEnUSInitCanvasJsonSchema, defaultZhCNInitCanvasJsonSchema)
}

// GetDefaultInitCanvasJsonSchemaChat 根据语言环境获取默认的聊天工作流画布配置
// 相比普通工作流，聊天工作流包含用户输入和会话管理等特定配置
func GetDefaultInitCanvasJsonSchemaChat(locale i18n.Locale, name string) string {
	return ternary.IFElse(locale == i18n.LocaleEN, fmt.Sprintf(defaultEnUSInitCanvasJsonSchemaChat, name), fmt.Sprintf(defaultZhCNInitCanvasJsonSchemaChat, name))
}
