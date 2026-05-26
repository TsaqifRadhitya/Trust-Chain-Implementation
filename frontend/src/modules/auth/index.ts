import { apiClient } from '../../lib/axios';
import { setAccessToken, setRefreshToken } from '../../lib/auth-storage';
import type { LoginResponse } from './type';

export async function loginUser(email: string, password: string): Promise<LoginResponse> {
  try {
    const response = await apiClient.post('/auth/login', { email, password });

    // Ambil data token dan informasi user dari body response
    const { token, refresh_token, user } = response.data.data;

    // Simpan token ke storage agar bisa digunakan oleh Axios Interceptor
    if (token) setAccessToken(token);
    if (refresh_token) setRefreshToken(refresh_token);

    return {
      success: true,
      user: {
        email: user.email,
        name: user.name,
      },
    };
  } catch (error: any) {
    const errorMessage = error.response?.data?.error || error.response?.data?.message || 'Login failed';
    return {
      success: false,
      user: null,
      error: errorMessage,
    };
  }
}
