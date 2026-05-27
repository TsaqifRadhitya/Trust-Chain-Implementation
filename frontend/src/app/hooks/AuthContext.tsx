import React, { createContext, useContext, useState, useCallback } from 'react';
import { loginUser } from '../../modules/auth';
import type { User } from '../../modules/auth/type';

interface AuthContextType {
  isAuthenticated: boolean;
  user: User | null;
  login: (email: string, password: string) => Promise<{ success: boolean; error?: string }>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [isAuthenticated, setIsAuthenticated] = useState(() => {
    return localStorage.getItem('tc_auth') === 'true';
  });
  const [user, setUser] = useState<User | null>(() => {
    const stored = localStorage.getItem('tc_user');
    return stored ? JSON.parse(stored) : null;
  });

  const login = useCallback(async (email: string, password: string) => {
    try {
      const result = await loginUser(email, password);
      if (result.success && result.user) {
        setIsAuthenticated(true);
        setUser(result.user);
        localStorage.setItem('tc_auth', 'true');
        localStorage.setItem('tc_user', JSON.stringify(result.user));
        return { success: true };
      }
      return { success: false, error: result.error || 'Login gagal' };
    } catch (error: unknown) {
      return { success: false, error: (error as Error).message || 'Login gagal' };
    }
  }, []);

  const logout = useCallback(() => {
    setIsAuthenticated(false);
    setUser(null);
    localStorage.removeItem('tc_auth');
    localStorage.removeItem('tc_user');
  }, []);


  return (
    <AuthContext.Provider value={{ isAuthenticated, user, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

// eslint-disable-next-line react-refresh/only-export-components
export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
