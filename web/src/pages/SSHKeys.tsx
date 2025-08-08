import React, { useEffect, useState } from 'react';
import { apiService } from '../services/api';
import { SSHKey } from '../types';
import { Key, Plus, Trash2, Copy } from 'lucide-react';

const SSHKeys: React.FC = () => {
  const [sshKeys, setSSHKeys] = useState<SSHKey[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadSSHKeys();
  }, []);

  const loadSSHKeys = async () => {
    try {
      const response = await apiService.getSSHKeys();
      if (response.success && response.data) {
        setSSHKeys(response.data);
      }
    } catch (error) {
      console.error('Failed to load SSH keys:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteKey = async (id: number) => {
    if (confirm('确定要删除这个SSH密钥吗？')) {
      try {
        await apiService.deleteSSHKey(id);
        setSSHKeys(sshKeys.filter(key => key.id !== id));
      } catch (error) {
        console.error('Failed to delete SSH key:', error);
      }
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    // 可以添加复制成功提示
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
          <h1 className="text-2xl font-bold text-gray-900">SSH密钥管理</h1>
          <p className="mt-1 text-sm text-gray-500">
            管理用于Git仓库访问的SSH密钥
          </p>
        </div>
        <button className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700">
          <Plus className="h-4 w-4 mr-2" />
          添加SSH密钥
        </button>
      </div>

      {sshKeys.length === 0 ? (
        <div className="text-center py-12">
          <Key className="mx-auto h-12 w-12 text-gray-400" />
          <h3 className="mt-2 text-sm font-medium text-gray-900">暂无SSH密钥</h3>
          <p className="mt-1 text-sm text-gray-500">
            添加SSH密钥以访问私有Git仓库
          </p>
          <div className="mt-6">
            <button className="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700">
              <Plus className="h-4 w-4 mr-2" />
              添加SSH密钥
            </button>
          </div>
        </div>
      ) : (
        <div className="bg-white shadow overflow-hidden sm:rounded-md">
          <ul className="divide-y divide-gray-200">
            {sshKeys.map((sshKey) => (
              <li key={sshKey.id}>
                <div className="px-4 py-4">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center">
                      <div className="flex-shrink-0">
                        <Key className="h-6 w-6 text-gray-400" />
                      </div>
                      <div className="ml-4">
                        <div className="flex items-center">
                          <p className="text-sm font-medium text-indigo-600 truncate">
                            {sshKey.name}
                          </p>
                        </div>
                        <div className="mt-2">
                          <p className="text-sm text-gray-500">
                            创建时间: {new Date(sshKey.created_at).toLocaleString()}
                          </p>
                        </div>
                      </div>
                    </div>
                    <div className="flex items-center space-x-2">
                      <button 
                        onClick={() => copyToClipboard(sshKey.public_key)}
                        className="text-indigo-600 hover:text-indigo-900"
                        title="复制公钥"
                      >
                        <Copy className="h-4 w-4" />
                      </button>
                      <button 
                        onClick={() => handleDeleteKey(sshKey.id)}
                        className="text-red-600 hover:text-red-900"
                      >
                        <Trash2 className="h-4 w-4" />
                      </button>
                    </div>
                  </div>
                  <div className="mt-4">
                    <div className="bg-gray-50 rounded-md p-3">
                      <p className="text-xs font-mono text-gray-600 break-all">
                        {sshKey.public_key.substring(0, 100)}...
                      </p>
                    </div>
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

export default SSHKeys;