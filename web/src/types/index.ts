export interface User {
  id: number;
  username: string;
  email: string;
  role: 'admin' | 'user';
  created_at: string;
  updated_at: string;
}

export interface Project {
  id: number;
  name: string;
  description: string;
  repository_url: string;
  branch: string;
  build_command: string;
  deploy_command: string;
  environment_variables: Record<string, string>;
  ssh_key_id?: number;
  user_id: number;
  created_at: string;
  updated_at: string;
}

export interface Pipeline {
  id: number;
  name: string;
  description: string;
  project_id: number;
  trigger_type: 'manual' | 'webhook' | 'schedule';
  cron_expression?: string;
  steps: PipelineStep[];
  user_id: number;
  created_at: string;
  updated_at: string;
}

export interface PipelineStep {
  id: number;
  name: string;
  type: 'script' | 'deploy' | 'git_pull';
  command: string;
  order: number;
  pipeline_id: number;
}

export interface Deployment {
  id: number;
  pipeline_id: number;
  status: 'pending' | 'running' | 'success' | 'failed';
  logs: string;
  started_at?: string;
  finished_at?: string;
  user_id: number;
  created_at: string;
  updated_at: string;
}

export interface SSHKey {
  id: number;
  name: string;
  public_key: string;
  private_key: string;
  user_id: number;
  created_at: string;
  updated_at: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  user: User;
}

export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  message?: string;
  error?: string;
}