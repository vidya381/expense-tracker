'use client';

import { useState } from "react";
import { useRouter } from "next/navigation";

type ApiResponse = {
    success: boolean;
    message?: string;
    error?: string;
};

export default function RegisterPage() {
    const router = useRouter();
    const [username, setUsername] = useState("");
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [error, setError] = useState<string | null>(null);
    const [success, setSuccess] = useState<string | null>(null);
    const [loading, setLoading] = useState(false);

    function validateEmailFormat(email: string): boolean {
        return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
    }

    async function handleRegister(event: React.FormEvent) {
        event.preventDefault();
        setError(null);
        setSuccess(null);

        // Frontend validation
        if (!username.trim()) {
            setError("Username is required");
            return;
        }
        if (!email.trim()) {
            setError("Email is required");
            return;
        }
        if (!validateEmailFormat(email)) {
            setError("Please enter a valid email address");
            return;
        }
        if (password.length < 4) {
            setError("Password must be at least 4 characters");
            return;
        }

        setLoading(true);

        try {
            const formData = new FormData();
            formData.append("username", username);
            formData.append("email", email);
            formData.append("password", password);

            const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/register`, {
                method: "POST",
                body: formData,
            });

            const data: ApiResponse = await res.json();

            if (!data.success) {
                setError(data.error || "Registration failed");
                setLoading(false);
                return;
            }

            setSuccess(data.message || "Registered successfully!");
            setLoading(false);

            setTimeout(() => {
                router.push("/login");
            }, 2000);
        } catch {
            setError("Network error. Please try again.");
            setLoading(false);
        }
    }

    return (
        <main className="flex min-h-screen items-center justify-center bg-gray-50">
            <form
                onSubmit={handleRegister}
                className="bg-white shadow-md rounded px-8 pt-6 pb-8 mb-4 w-96"
            >
                <h2 className="mb-6 text-lg font-bold">Register</h2>

                {error && (
                    <div className="mb-4 text-red-600 bg-red-100 p-2 rounded">{error}</div>
                )}

                {success && (
                    <div className="mb-4 text-green-700 bg-green-100 p-2 rounded">{success}</div>
                )}

                <div className="mb-4">
                    <label className="block text-gray-700 mb-2">Username</label>
                    <input
                        type="text"
                        className="input input-bordered w-full"
                        value={username}
                        onChange={(e) => setUsername(e.target.value)}
                        disabled={loading}
                    />
                </div>

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
                    {loading ? "Registering..." : "Register"}
                </button>

                <p className="mt-4 text-center text-sm">
                    Already have an account?{" "}
                    <a href="/login" className="text-blue-600 underline">
                        Login
                    </a>
                </p>
            </form>
        </main>
    );
}
