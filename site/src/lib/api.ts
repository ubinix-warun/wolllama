const API_BASE = import.meta.env.VITE_API_URL || "";

interface ApiError {
  error: string;
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}/api${path}`, {
    credentials: "include",
    headers: { "Content-Type": "application/json", ...options?.headers },
    ...options,
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({})) as ApiError;
    throw new Error(body.error || `API error: ${res.status}`);
  }

  return res.json() as Promise<T>;
}

export interface Model {
  id: number;
  submitter_id: number;
  submitter_name?: string;
  avatar_url?: string;
  manifest_obj_id: string;
  display_name: string;
  description_md?: string;
  original_name?: string;
  tag?: string;
  total_size?: number;
  blob_count?: number;
  manifest_json?: string;
  available: boolean;
  created_at: string;
}

export interface User {
  id: number;
  github_id: number;
  username: string;
  avatar_url?: string;
  created_at: string;
}

interface ModelsResponse {
  models: Model[];
  offset: number;
  limit: number;
}

// Public endpoints
export function listModels(offset = 0, limit = 20, search = "") {
  const params = new URLSearchParams({ offset: String(offset), limit: String(limit) });
  if (search) params.set("search", search);
  return request<ModelsResponse>(`/models?${params}`);
}

export function getModel(id: number) {
  return request<Model>(`/models/${id}`);
}

export function getUserModels(userId: number) {
  return request<{ models: Model[] }>(`/users/${userId}/models`);
}

// Auth
export function getMe() {
  return request<User>("/auth/me");
}

export function getLoginUrl() {
  return `${API_BASE}/api/auth/login`;
}

// Authenticated
export function submitModel(manifestObjId: string, displayName: string, descriptionMd?: string) {
  return request<Model>("/models", {
    method: "POST",
    body: JSON.stringify({
      manifest_obj_id: manifestObjId,
      display_name: displayName,
      description_md: descriptionMd || null,
    }),
  });
}
