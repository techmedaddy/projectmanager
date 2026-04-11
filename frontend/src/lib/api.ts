export const API_BASE_URL =
  (import.meta.env.VITE_API_BASE_URL as string | undefined) || 'http://localhost:8080';

export class ApiError extends Error {
  constructor(
    public status: number,
    public message: string,
    public fields?: Record<string, string>
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

async function handleResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    let errorData;
    try {
      errorData = await response.json();
    } catch {
      throw new ApiError(response.status, 'An unexpected error occurred');
    }
    throw new ApiError(
      response.status,
      errorData.error || 'An unexpected error occurred',
      errorData.fields
    );
  }
  
  if (response.status === 204) {
    return {} as T;
  }
  
  return response.json();
}

export async function fetchApi<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const token = localStorage.getItem('taskflow_token');
  
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    ...options.headers,
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    headers,
  });

  return handleResponse<T>(response);
}
