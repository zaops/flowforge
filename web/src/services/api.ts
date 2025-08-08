import { User, Project, Pipeline, Deployment, SSHKey, LoginRequest, LoginResponse, ApiResponse } from '@/types';

const API_BASE_URL = '/api';

class ApiService {
  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
    const token = localStorage.getItem('token');
    
    const config: RequestInit = {
      headers: {
        'Content-Type': 'application/json',
        ...(token && { Authorization: `Bearer ${token}` }),
        ...options.headers,
      },
      ...options,
    };

    try {
      const response = await fetch(`${API_BASE_URL}${endpoint}`, config);
      const data = await response.json();
      
      if (!response.ok) {
        throw new Error(data.message || 'Request failed');
      }
      
      return data;
    } catch (error) {
      console.error('API request failed:', error);
      throw error;
    }
  }

  // Auth APIs
  async login(credentials: LoginRequest): Promise<ApiResponse<LoginResponse>> {
    return this.request<LoginResponse>('/auth/login', {
      method: 'POST',
      body: JSON.stringify(credentials),
    });
  }

  async logout(): Promise<ApiResponse> {
    return this.request('/auth/logout', {
      method: 'POST',
    });
  }

  async getCurrentUser(): Promise<ApiResponse<User>> {
    return this.request<User>('/auth/me');
  }

  // User APIs
  async getUsers(): Promise<ApiResponse<User[]>> {
    return this.request<User[]>('/users');
  }

  async createUser(user: Partial<User>): Promise<ApiResponse<User>> {
    return this.request<User>('/users', {
      method: 'POST',
      body: JSON.stringify(user),
    });
  }

  async updateUser(id: number, user: Partial<User>): Promise<ApiResponse<User>> {
    return this.request<User>(`/users/${id}`, {
      method: 'PUT',
      body: JSON.stringify(user),
    });
  }

  async deleteUser(id: number): Promise<ApiResponse> {
    return this.request(`/users/${id}`, {
      method: 'DELETE',
    });
  }

  // Project APIs
  async getProjects(): Promise<ApiResponse<Project[]>> {
    return this.request<Project[]>('/projects');
  }

  async getProject(id: number): Promise<ApiResponse<Project>> {
    return this.request<Project>(`/projects/${id}`);
  }

  async createProject(project: Partial<Project>): Promise<ApiResponse<Project>> {
    return this.request<Project>('/projects', {
      method: 'POST',
      body: JSON.stringify(project),
    });
  }

  async updateProject(id: number, project: Partial<Project>): Promise<ApiResponse<Project>> {
    return this.request<Project>(`/projects/${id}`, {
      method: 'PUT',
      body: JSON.stringify(project),
    });
  }

  async deleteProject(id: number): Promise<ApiResponse> {
    return this.request(`/projects/${id}`, {
      method: 'DELETE',
    });
  }

  // Pipeline APIs
  async getPipelines(projectId?: number): Promise<ApiResponse<Pipeline[]>> {
    const query = projectId ? `?project_id=${projectId}` : '';
    return this.request<Pipeline[]>(`/pipelines${query}`);
  }

  async getPipeline(id: number): Promise<ApiResponse<Pipeline>> {
    return this.request<Pipeline>(`/pipelines/${id}`);
  }

  async createPipeline(pipeline: Partial<Pipeline>): Promise<ApiResponse<Pipeline>> {
    return this.request<Pipeline>('/pipelines', {
      method: 'POST',
      body: JSON.stringify(pipeline),
    });
  }

  async updatePipeline(id: number, pipeline: Partial<Pipeline>): Promise<ApiResponse<Pipeline>> {
    return this.request<Pipeline>(`/pipelines/${id}`, {
      method: 'PUT',
      body: JSON.stringify(pipeline),
    });
  }

  async deletePipeline(id: number): Promise<ApiResponse> {
    return this.request(`/pipelines/${id}`, {
      method: 'DELETE',
    });
  }

  async runPipeline(id: number): Promise<ApiResponse<Deployment>> {
    return this.request<Deployment>(`/pipelines/${id}/run`, {
      method: 'POST',
    });
  }

  // Deployment APIs
  async getDeployments(pipelineId?: number): Promise<ApiResponse<Deployment[]>> {
    const query = pipelineId ? `?pipeline_id=${pipelineId}` : '';
    return this.request<Deployment[]>(`/deployments${query}`);
  }

  async getDeployment(id: number): Promise<ApiResponse<Deployment>> {
    return this.request<Deployment>(`/deployments/${id}`);
  }

  async getDeploymentLogs(id: number): Promise<ApiResponse<string>> {
    return this.request<string>(`/deployments/${id}/logs`);
  }

  // SSH Key APIs
  async getSSHKeys(): Promise<ApiResponse<SSHKey[]>> {
    return this.request<SSHKey[]>('/ssh-keys');
  }

  async createSSHKey(sshKey: Partial<SSHKey>): Promise<ApiResponse<SSHKey>> {
    return this.request<SSHKey>('/ssh-keys', {
      method: 'POST',
      body: JSON.stringify(sshKey),
    });
  }

  async deleteSSHKey(id: number): Promise<ApiResponse> {
    return this.request(`/ssh-keys/${id}`, {
      method: 'DELETE',
    });
  }
}

export const apiService = new ApiService();
export default apiService;