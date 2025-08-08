import React, { useState } from 'react';
import { Settings as SettingsIcon, Save, Database, Server, Mail, Shield } from 'lucide-react';

const Settings: React.FC = () => {
  const [settings, setSettings] = useState({
    // 系统设置
    systemName: 'FlowForge',
    systemDescription: '可视化部署工具',
    maxConcurrentDeployments: 5,
    deploymentTimeout: 3600,
    
    // 数据库设置
    dbHost: 'localhost',
    dbPort: '5432',
    dbName: 'flowforge',
    
    // 邮件设置
    smtpHost: '',
    smtpPort: '587',
    smtpUser: '',
    smtpPassword: '',
    
    // 安全设置
    jwtSecret: '***hidden***',
    sessionTimeout: 24,
    enableTwoFactor: false,
  });

  const handleSave = () => {
    // 保存设置逻辑
    console.log('Saving settings:', settings);
  };

  const handleInputChange = (key: string, value: string | number | boolean) => {
    setSettings(prev => ({
      ...prev,
      [key]: value
    }));
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">系统设置</h1>
        <p className="mt-1 text-sm text-gray-500">
          配置系统参数和集成设置
        </p>
      </div>

      <div className="space-y-6">
        {/* 基本设置 */}
        <div className="bg-white shadow rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <div className="flex items-center mb-4">
              <SettingsIcon className="h-5 w-5 text-gray-400 mr-2" />
              <h3 className="text-lg leading-6 font-medium text-gray-900">
                基本设置
              </h3>
            </div>
            <div className="grid grid-cols-1 gap-6 sm:grid-cols-2">
              <div>
                <label className="block text-sm font-medium text-gray-700">
                  系统名称
                </label>
                <input
                  type="text"
                  value={settings.systemName}
                  onChange={(e) => handleInputChange('systemName', e.target.value)}
                  className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">
                  系统描述
                </label>
                <input
                  type="text"
                  value={settings.systemDescription}
                  onChange={(e) => handleInputChange('systemDescription', e.target.value)}
                  className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">
                  最大并发部署数
                </label>
                <input
                  type="number"
                  value={settings.maxConcurrentDeployments}
                  onChange={(e) => handleInputChange('maxConcurrentDeployments', parseInt(e.target.value))}
                  className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">
                  部署超时时间 (秒)
                </label>
                <input
                  type="number"
                  value={settings.deploymentTimeout}
                  onChange={(e) => handleInputChange('deploymentTimeout', parseInt(e.target.value))}
                  className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                />
              </div>
            </div>
          </div>
        </div>

        {/* 数据库设置 */}
        <div className="bg-white shadow rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <div className="flex items-center mb-4">
              <Database className="h-5 w-5 text-gray-400 mr-2" />
              <h3 className="text-lg leading-6 font-medium text-gray-900">
                数据库设置
              </h3>
            </div>
            <div className="grid grid-cols-1 gap-6 sm:grid-cols-2">
              <div>
                <label className="block text-sm font-medium text-gray-700">
                  数据库主机
                </label>
                <input
                  type="text"
                  value={settings.dbHost}
                  onChange={(e) => handleInputChange('dbHost', e.target.value)}
                  className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">
                  数据库端口
                </label>
                <input
                  type="text"
                  value={settings.dbPort}
                  onChange={(e) => handleInputChange('dbPort', e.target.value)}
                  className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                />
              </div>
              <div className="sm:col-span-2">
                <label className="block text-sm font-medium text-gray-700">
                  数据库名称
                </label>
                <input
                  type="text"
                  value={settings.dbName}
                  onChange={(e) => handleInputChange('dbName', e.target.value)}
                  className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                />
              </div>
            </div>
          </div>
        </div>

        {/* 邮件设置 */}
        <div className="bg-white shadow rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <div className="flex items-center mb-4">
              <Mail className="h-5 w-5 text-gray-400 mr-2" />
              <h3 className="text-lg leading-6 font-medium text-gray-900">
                邮件设置
              </h3>
            </div>
            <div className="grid grid-cols-1 gap-6 sm:grid-cols-2">
              <div>
                <label className="block text-sm font-medium text-gray-700">
                  SMTP主机
                </label>
                <input
                  type="text"
                  value={settings.smtpHost}
                  onChange={(e) => handleInputChange('smtpHost', e.target.value)}
                  className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">
                  SMTP端口
                </label>
                <input
                  type="text"
                  value={settings.smtpPort}
                  onChange={(e) => handleInputChange('smtpPort', e.target.value)}
                  className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">
                  SMTP用户名
                </label>
                <input
                  type="text"
                  value={settings.smtpUser}
                  onChange={(e) => handleInputChange('smtpUser', e.target.value)}
                  className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">
                  SMTP密码
                </label>
                <input
                  type="password"
                  value={settings.smtpPassword}
                  onChange={(e) => handleInputChange('smtpPassword', e.target.value)}
                  className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                />
              </div>
            </div>
          </div>
        </div>

        {/* 安全设置 */}
        <div className="bg-white shadow rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <div className="flex items-center mb-4">
              <Shield className="h-5 w-5 text-gray-400 mr-2" />
              <h3 className="text-lg leading-6 font-medium text-gray-900">
                安全设置
              </h3>
            </div>
            <div className="grid grid-cols-1 gap-6 sm:grid-cols-2">
              <div>
                <label className="block text-sm font-medium text-gray-700">
                  JWT密钥
                </label>
                <input
                  type="password"
                  value={settings.jwtSecret}
                  onChange={(e) => handleInputChange('jwtSecret', e.target.value)}
                  className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">
                  会话超时时间 (小时)
                </label>
                <input
                  type="number"
                  value={settings.sessionTimeout}
                  onChange={(e) => handleInputChange('sessionTimeout', parseInt(e.target.value))}
                  className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                />
              </div>
              <div className="sm:col-span-2">
                <div className="flex items-center">
                  <input
                    type="checkbox"
                    checked={settings.enableTwoFactor}
                    onChange={(e) => handleInputChange('enableTwoFactor', e.target.checked)}
                    className="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded"
                  />
                  <label className="ml-2 block text-sm text-gray-900">
                    启用双因素认证
                  </label>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* 保存按钮 */}
        <div className="flex justify-end">
          <button
            onClick={handleSave}
            className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700"
          >
            <Save className="h-4 w-4 mr-2" />
            保存设置
          </button>
        </div>
      </div>
    </div>
  );
};

export default Settings;