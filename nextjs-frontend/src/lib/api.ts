import { useAuth } from '../context/AuthContext';

// A regular function to call API with token, independent of React hooks
export async function apiFetch(
    url: string,
    options: RequestInit = {}
): Promise<any> {
    // Extract token from localStorage
    const token = typeof window !== 'undefined' ? localStorage.getItem('jwt_token') : null;

    const headers = new Headers(options.headers);
    headers.set('Content-Type', 'application/json');
    if (token) {
        headers.set('Authorization', `Bearer ${token}`);
    }

    const res = await fetch(url, options);

    if (!res.ok) {
        if (res.status === 401) {
            localStorage.removeItem('jwt_token');
            window.location.href = '/login';
            return Promise.reject(new Error('Unauthorized — please log in again'));
        }
        const errorBody = await res.json().catch(() => ({}));
        throw new Error(errorBody.error || 'API request failed');
    }
    return res.json();
}

export function useApi() {
    const { token } = useAuth();

    return async function (url: string, options: RequestInit = {}) {
        const headers = new Headers(options.headers);
        headers.set('Content-Type', 'application/json');
        if (token) {
            headers.set('Authorization', `Bearer ${token}`);
        }

        const res = await fetch(url, {
            ...options,
            headers,
        });

        if (!res.ok) {
            const errorBody = await res.json().catch(() => ({}));
            throw new Error(errorBody.error || 'API request failed');
        }
        return res.json();
    };
}