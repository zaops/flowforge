import React, { useEffect, useState } from 'react';
import { apiService } from '../services/api';
import { Deployment } from '../types';
import { Rocket, CheckCircle, XCircle, Clock, Play, Eye } from 'lucide-react';

const Deployments: React.FC = () => {
  const [deployments, setDeployments] = useState<Deployment[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadDeployments();
  }, []);

  const loadDeployments = async () => {
    try {
      const response = await apiService.getDeployments();
      if (response.success && response.data) {
        setDeployments(response.data);
      }
    } catch (error) {
      console.error('Failed to load deployments:', error);
    } finally {
      setLoading(false);
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'success':
        return <CheckCircle className="h-5 w-5 text-green-500" />;
      case 'failed':
        return <XCircle className="h-5 w-5 text-red-500" />;
      case 'running':
        return <Play className="h-5 w-5 text-blue-500" />;
      default:
        return <Clock className="h-5 w-5 text-gray-500" />;
    }
  };

  const getStatusText = (status: string) => {
    switch (status) {
      case 'success':
        return '成功';
      case 'failed':
        return '失败';
      case 'running':
        return '运行中';
      default:
        return '等待中';
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'success':
        return 'bg-green-100 text-green-800';
      case 'failed':
        return 'bg-red-100 text-red-800';
      case 'running':
        return 'bg-blue-100 text-blue-800';
      default:
        return 'bg-gray-100 text-gray-800';
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
      <div>
        <h1 className="text-2xl font-bold text-gray-900">部署记录</h1>
        <p className="mt-1 text-sm text-gray-500">
          查看所有部署历史和状态
        </p>
      </div>

      {deployments.length === 0 ? (
        <div className="text-center py-12">
          <Rocket className="mx-auto h-12 w-12 text-gray-400" />
          <h3 className="mt-2 text-sm font-medium text-gray-900">暂无部署记录</h3>
          <p className="mt-1 text-sm text-gray-500">
            当您运行流水线时，部署记录将显示在这里
          </p>
        </div>
      ) : (
        <div className="bg-white shadow overflow-hidden sm:rounded-md">
          <ul className="divide-y divide-gray-200">
            {deployments.map((deployment) => (
              <li key={deployment.id}>
                <div className="px-4 py-4 flex items-center justify-between">
                  <div className="flex items-center">
                    <div className="flex-shrink-0">
                      {getStatusIcon(deployment.status)}
                    </div>
                    <div className="ml-4">
                      <div className="flex items-center">
                        <p className="text-sm font-medium text-gray-900">
                          部署 #{deployment.id}
                        </p>
                        <span className={`ml-2 inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getStatusColor(deployment.status)}`}>
                          {getStatusText(deployment.status)}
                        </span>
                      </div>
                      <div className="mt-2 flex items-center text-sm text-gray-500">
                        <p>流水线 ID: {deployment.pipeline_id}</p>
                        <span className="mx-2">•</span>
                        <p>{new Date(deployment.created_at).toLocaleString()}</p>
                        {deployment.finished_at && (
                          <>
                            <span className="mx-2">•</span>
                            <p>耗时: {Math.round((new Date(deployment.finished_at).getTime() - new Date(deployment.started_at || deployment.created_at).getTime()) / 1000)}秒</p>
                          </>
                        )}
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center space-x-2">
                    <button className="text-indigo-600 hover:text-indigo-900">
                      <Eye className="h-4 w-4" />
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

export default Deployments;