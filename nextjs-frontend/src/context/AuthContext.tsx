'use client';

import React, {
    createContext,
    useContext,
    useState,
    useEffect,
    ReactNode,
} from 'react';
import { useRouter } from 'next/navigation';

interface AuthContextType {
    token: string | null;
    setToken: (token: string | null) => void;
    logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
    const [token, setToken] = useState<string | null>(null);
    const router = useRouter();

    // Load token from localStorage on mount
    useEffect(() => {
        try {
            const storedToken = localStorage.getItem('jwt_token');
            if (storedToken) setToken(storedToken);
        } catch (e) {
            // localStorage possibly not available or error occurred
            console.error('Failed to get token from localStorage', e);
        }
    }, []);

    // Set token in localStorage when it changes
    useEffect(() => {
        try {
            if (token) {
                localStorage.setItem('jwt_token', token);
            } else {
                localStorage.removeItem('jwt_token');
            }
        } catch (e) {
            console.error('Failed to update token in localStorage', e);
        }
    }, [token]);

    function logout() {
        setToken(null);
        localStorage.removeItem('jwt_token');
        router.push('/login');  // Redirect after logout
    }

    return (
        <AuthContext.Provider value={{ token, setToken, logout }}>
            {children}
        </AuthContext.Provider>
    );
}

export function useAuth() {
    const context = useContext(AuthContext);
    if (!context) {
        throw new Error('useAuth must be used within an AuthProvider');
    }
    return context;
}
