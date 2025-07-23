'use client';

import { useState } from "react";
import { useRouter } from "next/navigation";

type ApiResponse = {
    success: boolean;
    token?: string;
    error?: string;
};

export default function LoginPage() {
    const router = useRouter();
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [error, setError] = useState<string | null>(null);
    const [loading, setLoading] = useState(false);

    function validateEmailFormat(email: string): boolean {
        // email format check
        return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
    }

    async function handleLogin(event: React.FormEvent) {
        event.preventDefault();
        setError(null);

        // Frontend validation
        if (!email.trim()) {
            setError("Email is required");
            return;
        }
        if (!validateEmailFormat(email)) {
            setError("Please enter a valid email address");
            return;
        }
        if (!password) {
            setError("Password is required");
            return;
        }

        setLoading(true);

        try {
            const formData = new FormData();
            formData.append("email", email);
            formData.append("password", password);

            const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/login`, {
                method: "POST",
                body: formData,
            });

            const data: ApiResponse = await res.json();

            if (!data.success) {
                setError(data.error || "Login failed");
                setLoading(false);
                return;
            }

            if (data.token) {
                localStorage.setItem("jwt_token", data.token);
                router.push("/"); // Redirect to home
            } else {
                setError("No token received from server");
                setLoading(false);
            }
        } catch {
            setError("Network error. Please try again.");
            setLoading(false);
        }
    }

    return (
        <main className="flex min-h-screen items-center justify-center bg-gray-50">
            <form
                onSubmit={handleLogin}
                className="bg-white shadow-md rounded px-8 pt-6 pb-8 mb-4 w-96"
            >
                <h2 className="mb-6 text-lg font-bold">Login</h2>
                {error && (
                    <div className="mb-4 text-red-600 bg-red-100 p-2 rounded">{error}</div>
                )}

                <div className="mb-4">
                    <label className="block text-gray-700 mb-2">Email</label>
                    <input
                        type="email"
                        className="input input-bordered w-full"
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
                        disabled={loading}
                    />
                </div>

                <div className="mb-6">
                    <label className="block text-gray-700 mb-2">Password</label>
                    <input
                        type="password"
                        className="input input-bordered w-full"
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        disabled={loading}
                    />
                </div>

                <button
                    type="submit"
                    disabled={loading}
                    className={`w-full py-2 px-4 rounded font-bold text-white ${loading ? "bg-blue-400" : "bg-blue-600 hover:bg-blue-700"
                        }`}
                >
                    {loading ? "Logging in..." : "Login"}
                </button>

                <p className="mt-4 text-center text-sm">
                    Don't have an account?{" "}
                    <a href="/register" className="text-blue-600 underline">
                        Register
                    </a>
                </p>
            </form>
        </main>
    );
}
