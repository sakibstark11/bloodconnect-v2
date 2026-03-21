import {
  type ExtendedDonationRequest,
  type ListRequestsResponse,
  type LoginResponse,
  type NotificationsResponse,
  type SignupRequest,
  type User
} from './types';

const getApiBaseUrl = () => {
  return import.meta.env.VITE_API_URL || 'http://localhost:8080';
};

export const API_BASE_URL = getApiBaseUrl();

const fetchWithToken = async (path: string, token: string, options: RequestInit = {}) => {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...options,
    headers: {
      ...options.headers,
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
  });

  if (!response.ok) {
    if (response.status === 401) {
      // Could trigger logout here if needed
    }
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || `Request failed with status ${response.status}`);
  }

  if (response.status === 204) return null;
  return response.json();
};

export const api = {
  auth: {
    login: async (email: string, password: string): Promise<LoginResponse> => {
      const response = await fetch(`${API_BASE_URL}/users/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
      });
      if (!response.ok) throw new Error('Login failed');
      return response.json();
    },
    signup: async (data: SignupRequest): Promise<{ id: string }> => {
      const response = await fetch(`${API_BASE_URL}/users/signup`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
      });
      if (!response.ok) throw new Error('Signup failed');
      return response.json();
    },
    getMe: (token: string): Promise<User> => fetchWithToken('/users/me', token),
  },
  requests: {
    list: (token: string, filters: Record<string, string> = {}): Promise<ListRequestsResponse> => {
      const params = new URLSearchParams(filters).toString();
      return fetchWithToken(`/requests${params ? `?${params}` : ''}`, token);
    },
    get: (token: string, id: string): Promise<ExtendedDonationRequest> =>
      fetchWithToken(`/requests/${id}`, token),
    create: (token: string, data: any): Promise<{ id: string }> =>
      fetchWithToken('/requests', token, { method: 'POST', body: JSON.stringify(data) }),
    respond: (token: string, id: string, action: string): Promise<void> =>
      fetchWithToken(`/requests/${id}/respond`, token, { method: 'POST', body: JSON.stringify({ action }) }),
    cancel: (token: string, id: string): Promise<void> =>
      fetchWithToken(`/requests/${id}/cancel`, token, { method: 'POST' }),
  },
  notifications: {
    list: (token: string, lastId?: string): Promise<NotificationsResponse> => {
      const path = `/notifications${lastId ? `?last_notification_id=${lastId}` : ''}`;
      return fetchWithToken(path, token);
    },
  },
  user: {
    updateHealth: (token: string, data: { info_type: string; details: string }): Promise<void> =>
      fetchWithToken('/users/me/health', token, { method: 'PUT', body: JSON.stringify(data) }),
    updateLocation: (token: string, data: { lat: number; lng: number }): Promise<void> =>
      fetchWithToken('/users/me/location', token, { method: 'PUT', body: JSON.stringify(data) }),
  }
};
