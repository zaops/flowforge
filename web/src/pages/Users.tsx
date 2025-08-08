import React, { useEffect, useState } from 'react';
import { apiService } from '../services/api';
import { User } from '../types';
import { Users as UsersIcon, Plus, Edit, Trash2, Shield, User as UserIcon } from 'lucide-react';

const Users: React.FC = () => {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadUsers();
  }, []);

  const loadUsers = async () => {
    try {
      const response = await apiService.getUsers();
      if (response.success && response.data) {
        setUsers(response.data);
      }
    } catch (error) {
      console.error('Failed to load users:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteUser = async (id: number) => {
    if (confirm('确定要删除这个用户吗？')) {
      try {
        await apiService.deleteUser(id);
        setUsers(users.filter(user => user.id !== id));
      } catch (error) {
        console.error('Failed to delete user:', error);
      }
    }
  };

  const getRoleIcon = (role: string) => {
    return role === 'admin' ? (
      <Shield className="h-4 w-4 text-red-500" />
    ) : (
      <UserIcon className="h-4 w-4 text-gray-500" />
    );
  };

  const getRoleText = (role: string) => {
    return role === 'admin' ? '管理员' : '普通用户';
  };

  const getRoleColor = (role: string) => {
    return role === 'admin' 
      ? 'bg-red-100 text-red-800' 
      : 'bg-gray-100 text-gray-800';
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
          <h1 className="text-2xl font-bold text-gray-900">用户管理</h1>
          <p className="mt-1 text-sm text-gray-500">
            管理系统用户和权限
          </p>
        </div>
        <button className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700">
          <Plus className="h-4 w-4 mr-2" />
          添加用户
        </button>
      </div>

      {users.length === 0 ? (
        <div className="text-center py-12">
          <UsersIcon className="mx-auto h-12 w-12 text-gray-400" />
          <h3 className="mt-2 text-sm font-medium text-gray-900">暂无用户</h3>
          <p className="mt-1 text-sm text-gray-500">
            开始添加系统用户
          </p>
          <div className="mt-6">
            <button className="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700">
              <Plus className="h-4 w-4 mr-2" />
              添加用户
            </button>
          </div>
        </div>
      ) : (
        <div className="bg-white shadow overflow-hidden sm:rounded-md">
          <ul className="divide-y divide-gray-200">
            {users.map((user) => (
              <li key={user.id}>
                <div className="px-4 py-4 flex items-center justify-between">
                  <div className="flex items-center">
                    <div className="flex-shrink-0">
                      <div className="h-10 w-10 rounded-full bg-gray-200 flex items-center justify-center">
                        <UserIcon className="h-6 w-6 text-gray-500" />
                      </div>
                    </div>
                    <div className="ml-4">
                      <div className="flex items-center">
                        <p className="text-sm font-medium text-gray-900">
                          {user.username}
                        </p>
                        <span className={`ml-2 inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getRoleColor(user.role)}`}>
                          <span className="mr-1">{getRoleIcon(user.role)}</span>
                          {getRoleText(user.role)}
                        </span>
                      </div>
                      <div className="mt-2 flex items-center text-sm text-gray-500">
                        <p>{user.email}</p>
                        <span className="mx-2">•</span>
                        <p>创建时间: {new Date(user.created_at).toLocaleDateString()}</p>
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center space-x-2">
                    <button className="text-indigo-600 hover:text-indigo-900">
                      <Edit className="h-4 w-4" />
                    </button>
                    <button 
                      onClick={() => handleDeleteUser(user.id)}
                      className="text-red-600 hover:text-red-900"
                    >
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

export default Users;