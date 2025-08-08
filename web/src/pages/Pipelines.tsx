import React, { useEffect, useState } from 'react';
import { apiService } from '../services/api';
import { Pipeline } from '../types';
import { GitBranch, Plus, Play, Edit, Trash2 } from 'lucide-react';

const Pipelines: React.FC = () => {
  const [pipelines, setPipelines] = useState<Pipeline[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadPipelines();
  }, []);

  const loadPipelines = async () => {
    try {
      const response = await apiService.getPipelines();
      if (response.success && response.data) {
        setPipelines(response.data);
      }
    } catch (error) {
      console.error('Failed to load pipelines:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleRunPipeline = async (id: number) => {
    try {
      await apiService.runPipeline(id);
      // 可以添加成功提示
    } catch (error) {
      console.error('Failed to run pipeline:', error);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600"></div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">流水线管理</h1>
          <p className="mt-1 text-sm text-gray-500">
            管理您的CI/CD流水线
          </p>
        </div>
        <button className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700">
          <Plus className="h-4 w-4 mr-2" />
          新建流水线
        </button>
      </div>

      {pipelines.length === 0 ? (
        <div className="text-center py-12">
          <GitBranch className="mx-auto h-12 w-12 text-gray-400" />
          <h3 className="mt-2 text-sm font-medium text-gray-900">暂无流水线</h3>
          <p className="mt-1 text-sm text-gray-500">
            开始创建您的第一个流水线
          </p>
          <div className="mt-6">
            <button className="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700">
              <Plus className="h-4 w-4 mr-2" />
              新建流水线
            </button>
          </div>
        </div>
      ) : (
        <div className="bg-white shadow overflow-hidden sm:rounded-md">
          <ul className="divide-y divide-gray-200">
            {pipelines.map((pipeline) => (
              <li key={pipeline.id}>
                <div className="px-4 py-4 flex items-center justify-between">
                  <div className="flex items-center">
                    <div className="flex-shrink-0">
                      <GitBranch className="h-6 w-6 text-gray-400" />
                    </div>
                    <div className="ml-4">
                      <div className="flex items-center">
                        <p className="text-sm font-medium text-indigo-600 truncate">
                          {pipeline.name}
                        </p>
                        <span className="ml-2 inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
                          {pipeline.trigger_type}
                        </span>
                      </div>
                      <div className="mt-2">
                        <p className="text-sm text-gray-500">{pipeline.description}</p>
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center space-x-2">
                    <button 
                      onClick={() => handleRunPipeline(pipeline.id)}
                      className="text-green-600 hover:text-green-900"
                    >
                      <Play className="h-4 w-4" />
                    </button>
                    <button className="text-indigo-600 hover:text-indigo-900">
                      <Edit className="h-4 w-4" />
                    </button>
                    <button className="text-red-600 hover:text-red-900">
                      <Trash2 className="h-4 w-4" />
                    </button>
                  </div>
                </div>
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
};

export default Pipelines;