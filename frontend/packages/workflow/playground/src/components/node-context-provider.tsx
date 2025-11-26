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

/**
 * @file 节点上下文提供者组件
 * @description 为工作流节点提供统一的上下文环境，包括节点数据、错误状态和渲染场景
 */

import {
  startTransition,
  type PropsWithChildren,
  useState,
  useEffect,
} from 'react';

import { WorkflowNode, WorkflowNodeContext } from '@coze-workflow/base';
import {
  PlaygroundEntityContext,
  type FlowNodeEntity,
  FlowNodeFormData,
  FlowNodeErrorData,
} from '@flowgram-adapter/free-layout-editor';

import {
  NodeRenderSceneContext,
  type NodeRenderScene,
} from '@/contexts/node-render-context';

/**
 * 节点上下文提供者属性
 */
interface NodeContextProviderProps {
  node: FlowNodeEntity;
  scene?: NodeRenderScene;
}

/**
 * 节点上下文提供者组件
 *
 * 为工作流节点提供多层上下文：
 * - NodeRenderSceneContext: 渲染场景（编辑器/预览/测试等）
 * - WorkflowNodeContext: 工作流节点业务数据
 * - PlaygroundEntityContext: 画布节点实体
 */
export function NodeContextProvider({
  node,
  scene,
  children,
}: PropsWithChildren<NodeContextProviderProps>) {
  const workflowNode = useWorkflowNode(node);
  const [prevErrorMessage, setPrevErrorMessage] = useState<
    string | undefined
  >();

  useEffect(() => {
    if (!workflowNode.data) {
      return;
    }

    const errorMessage = workflowNode.registry?.checkError?.(
      workflowNode.data,
      node.context,
    );

    if (errorMessage !== prevErrorMessage) {
      if (errorMessage) {
        workflowNode.setError({
          name: 'CustomNodeError',
          message: errorMessage,
        });
      }
      setPrevErrorMessage(errorMessage);
    }
  }, [workflowNode]);

  return (
    <NodeRenderSceneContext.Provider value={scene}>
      <WorkflowNodeContext.Provider value={workflowNode}>
        <PlaygroundEntityContext.Provider value={node}>
          {children}
        </PlaygroundEntityContext.Provider>
      </WorkflowNodeContext.Provider>
    </NodeRenderSceneContext.Provider>
  );
}

/**
 * 监听节点数据变化并创建 WorkflowNode 实例
 *
 * 当底层节点数据变化时，自动更新业务层节点实例
 */
function useWorkflowNode(node: FlowNodeEntity) {
  const [workflowNode, setWorkflowNode] = useState<WorkflowNode>(
    new WorkflowNode(node),
  );

  // Monitor the underlying instance data changes and update the business layer instance
  useEffect(() => {
    const updateWorkflowNode = () => {
      startTransition(() => {
        const newWorkflowNode = new WorkflowNode(node);
        setWorkflowNode(newWorkflowNode);
      });
    };

    const dataChangeDisposer = node
      .getData(FlowNodeFormData)
      .onDataChange(() => updateWorkflowNode());

    const initialDisposer = node
      .getData(FlowNodeFormData)
      .formModel.onInitialized(() => updateWorkflowNode());

    const errorDisposer = node
      .getData<FlowNodeErrorData>(FlowNodeErrorData)
      .onDataChange(() => updateWorkflowNode());

    return () => {
      dataChangeDisposer?.dispose();
      initialDisposer?.dispose();
      errorDisposer?.dispose();
    };
  }, [node]);

  return workflowNode;
}
