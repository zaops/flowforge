import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { apiService } from '../services/api';
import { Project, Pipeline, Deployment } from '../types';
import {
  FolderOpen,
  GitBranch,
  Rocket,
  CheckCircle,
  XCircle,
  Clock,
  Play,
  TrendingUp,
  Activity
} from 'lucide-react';

const Dashboard: React.FC = () => {
  const [stats, setStats] = useState({
    projects: 0,
    pipelines: 0,
    deployments: 0,
    successRate: 0
  });
  const [recentDeployments, setRecentDeployments] = useState<Deployment[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadDashboardData();
  }, []);

  const loadDashboardData = async () => {
    try {
      const [projectsRes, pipelinesRes, deploymentsRes] = await Promise.all([
        apiService.getProjects(),
        apiService.getPipelines(),
        apiService.getDeployments()
      ]);

      if (projectsRes.success && pipelinesRes.success && deploymentsRes.success) {
        const projects = projectsRes.data || [];
        const pipelines = pipelinesRes.data || [];
        const deployments = deploymentsRes.data || [];

        const successCount = deployments.filter(d => d.status === 'success').length;
        const successRate = deployments.length > 0 ? (successCount / deployments.length) * 100 : 0;

        setStats({
          projects: projects.length,
          pipelines: pipelines.length,
          deployments: deployments.length,
          successRate: Math.round(successRate)
        });

        // 获取最近的部署记录
        const recent = deployments
          .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
          .slice(0, 5);
        setRecentDeployments(recent);
      }
    } catch (error) {
      console.error('Failed to load dashboard data:', error);
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
        <h1 className="text-2xl font-bold text-gray-900">仪表板</h1>
        <p className="mt-1 text-sm text-gray-500">
          查看项目概览和最近的部署活动
        </p>
      </div>

      {/* 统计卡片 */}
      <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4">
        <div className="bg-white overflow-hidden shadow rounded-lg">
          <div className="p-5">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <FolderOpen className="h-6 w-6 text-gray-400" />
              </div>
              <div className="ml-5 w-0 flex-1">
                <dl>
                  <dt className="text-sm font-medium text-gray-500 truncate">
                    项目总数
                  </dt>
                  <dd className="text-lg font-medium text-gray-900">
                    {stats.projects}
                  </dd>
                </dl>
              </div>
            </div>
          </div>
          <div className="bg-gray-50 px-5 py-3">
            <div className="text-sm">
              <Link
                to="/projects"
                className="font-medium text-indigo-700 hover:text-indigo-900"
              >
                查看所有项目
              </Link>
            </div>
          </div>
        </div>

        <div className="bg-white overflow-hidden shadow rounded-lg">
          <div className="p-5">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <GitBranch className="h-6 w-6 text-gray-400" />
              </div>
              <div className="ml-5 w-0 flex-1">
                <dl>
                  <dt className="text-sm font-medium text-gray-500 truncate">
                    流水线总数
                  </dt>
                  <dd className="text-lg font-medium text-gray-900">
                    {stats.pipelines}
                  </dd>
                </dl>
              </div>
            </div>
          </div>
          <div className="bg-gray-50 px-5 py-3">
            <div className="text-sm">
              <Link
                to="/pipelines"
                className="font-medium text-indigo-700 hover:text-indigo-900"
              >
                查看所有流水线
              </Link>
            </div>
          </div>
        </div>

        <div className="bg-white overflow-hidden shadow rounded-lg">
          <div className="p-5">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <Rocket className="h-6 w-6 text-gray-400" />
              </div>
              <div className="ml-5 w-0 flex-1">
                <dl>
                  <dt className="text-sm font-medium text-gray-500 truncate">
                    部署总数
                  </dt>
                  <dd className="text-lg font-medium text-gray-900">
                    {stats.deployments}
                  </dd>
                </dl>
              </div>
            </div>
          </div>
          <div className="bg-gray-50 px-5 py-3">
            <div className="text-sm">
              <Link
                to="/deployments"
                className="font-medium text-indigo-700 hover:text-indigo-900"
              >
                查看部署记录
              </Link>
            </div>
          </div>
        </div>

        <div className="bg-white overflow-hidden shadow rounded-lg">
          <div className="p-5">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <TrendingUp className="h-6 w-6 text-gray-400" />
              </div>
              <div className="ml-5 w-0 flex-1">
                <dl>
                  <dt className="text-sm font-medium text-gray-500 truncate">
                    成功率
                  </dt>
                  <dd className="text-lg font-medium text-gray-900">
                    {stats.successRate}%
                  </dd>
                </dl>
              </div>
            </div>
          </div>
          <div className="bg-gray-50 px-5 py-3">
            <div className="text-sm">
              <span className="font-medium text-gray-500">
                基于所有部署记录
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* 最近部署 */}
      <div className="bg-white shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg leading-6 font-medium text-gray-900">
              最近部署
            </h3>
            <Activity className="h-5 w-5 text-gray-400" />
          </div>
          
          {recentDeployments.length === 0 ? (
            <div className="text-center py-6">
              <p className="text-sm text-gray-500">暂无部署记录</p>
            </div>
          ) : (
            <div className="flow-root">
              <ul className="-mb-8">
                {recentDeployments.map((deployment, index) => (
                  <li key={deployment.id}>
                    <div className="relative pb-8">
                      {index !== recentDeployments.length - 1 && (
                        <span
                          className="absolute top-4 left-4 -ml-px h-full w-0.5 bg-gray-200"
                          aria-hidden="true"
                        />
                      )}
                      <div className="relative flex space-x-3">
                        <div>
                          {getStatusIcon(deployment.status)}
                        </div>
                        <div className="min-w-0 flex-1 pt-1.5 flex justify-between space-x-4">
                          <div>
                            <p className="text-sm text-gray-500">
                              部署 #{deployment.id} {getStatusText(deployment.status)}
                            </p>
                          </div>
                          <div className="text-right text-sm whitespace-nowrap text-gray-500">
                            {new Date(deployment.created_at).toLocaleString()}
                          </div>
                        </div>
                      </div>
                    </div>
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default Dashboard;