'use client';

import React, { useEffect, useState } from 'react';
import { useAuth } from '../context/AuthContext';
import { fetchCategories, addCategory, Category } from '../lib/categoryApi';

export default function CategoryManager() {
    const { token } = useAuth();
    const [categories, setCategories] = useState<Category[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    // Form state
    const [newName, setNewName] = useState('');
    const [newType, setNewType] = useState<'expense' | 'income'>('expense');
    const [adding, setAdding] = useState(false);

    useEffect(() => {
        if (!token) return;
        setLoading(true);
        fetchCategories(token)
            .then(data => setCategories(data))
            .catch(err => setError(err.message))
            .finally(() => setLoading(false));
    }, [token]);

    async function handleAddCategory(e: React.FormEvent) {
        e.preventDefault();
        setError(null);
        setAdding(true);

        try {
            await addCategory(token!, newName.trim(), newType);
            // Refresh categories
            const updated = await fetchCategories(token!);
            setCategories(updated);
            setNewName('');
            setNewType('expense');
        } catch (err: any) {
            setError(err.message);
        } finally {
            setAdding(false);
        }
    }

    return (
        <div className="max-w-md mx-auto p-4 bg-white shadow rounded">
            <h2 className="text-xl font-semibold mb-4">Your Categories</h2>

            {error && (
                <div className="mb-4 p-2 bg-red-100 text-red-700 rounded">{error}</div>
            )}

            {loading ? (
                <p>Loading categories...</p>
            ) : (
                <ul className="mb-6 max-h-48 overflow-y-auto border rounded p-2">
                    {categories.length === 0 ? (
                        <li className="text-gray-500">No categories added yet.</li>
                    ) : (
                        categories.map(cat => (
                            <li key={cat.id} className="py-1 border-b last:border-b-0 flex justify-between">
                                <span>{cat.name}</span>
                                <span className={`text-sm font-semibold ${cat.type === 'expense' ? 'text-red-600' : 'text-green-600'}`}>
                                    {cat.type}
                                </span>
                            </li>
                        ))
                    )}
                </ul>
            )}

            <form onSubmit={handleAddCategory} className="space-y-4">
                <div>
                    <label htmlFor="categoryName" className="block text-sm font-medium text-gray-700">
                        New Category Name
                    </label>
                    <input
                        id="categoryName"
                        type="text"
                        value={newName}
                        onChange={e => setNewName(e.target.value)}
                        disabled={adding}
                        required
                        className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded shadow-sm focus:outline-none focus:ring focus:border-blue-500 sm:text-sm"
                    />
                </div>
                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Type</label>
                    <select
                        value={newType}
                        onChange={e => setNewType(e.target.value as 'expense' | 'income')}
                        disabled={adding}
                        className="block w-full rounded border border-gray-300 px-3 py-2 shadow-sm focus:outline-none focus:ring focus:border-blue-500 sm:text-sm"
                    >
                        <option value="expense">Expense</option>
                        <option value="income">Income</option>
                    </select>
                </div>

                <button
                    type="submit"
                    disabled={adding}
                    className="w-full bg-blue-600 hover:bg-blue-700 text-white py-2 rounded font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                >
                    {adding ? 'Adding...' : 'Add Category'}
                </button>
            </form>
        </div>
    );
}
