import { useState } from 'react';
import type { ReactNode } from 'react';
import { AuthContext } from './context';

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [token, setToken] = useState<string | null>(null);
  const [userId, setUserId] = useState<string | null>(null);

  const login = (newToken: string) => {
    setToken(newToken);
    try {
      // Simple base64 decoding of the JWT payload
      const payload = JSON.parse(atob(newToken.split('.')[1]));
      setUserId(payload.user_id || payload.userId || payload.sub || null);
    } catch (e) {
      console.error('Failed to decode token', e);
      setUserId(null);
    }
  };

  const logout = () => {
    setToken(null);
    setUserId(null);
  };

  return (
    <AuthContext.Provider value={{ token, userId, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
};
