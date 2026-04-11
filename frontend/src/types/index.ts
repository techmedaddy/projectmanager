export interface User {
  id: string;
  name: string;
  email: string;
  created_at?: string;
}

export interface Project {
  id: string;
  name: string;
  description: string | null;
  owner_id: string;
  created_at: string;
}

export type TaskStatus = 'todo' | 'in_progress' | 'done';
export type TaskPriority = 'low' | 'medium' | 'high';

export interface Task {
  id: string;
  title: string;
  description: string | null;
  status: TaskStatus;
  priority: TaskPriority;
  project_id: string;
  assignee_id: string | null;
  creator_id: string;
  due_date: string | null;
  created_at: string;
  updated_at: string;
}

export interface AuthResponse {
  access_token: string;
}

export interface MeResponse {
  user: User;
}

export interface ProjectsResponse {
  projects: Project[];
}

export interface ProjectDetailResponse {
  project: Project;
  tasks: Task[];
}

export interface TasksResponse {
  tasks: Task[];
}

export interface AssigneesResponse {
  users: User[];
}
