'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';

interface Transaction {
    id: number;
    category_id: number;
    amount: number;
    description: string;
    date: string;
}

interface Category {
    id: number;
    name: string;
    type: 'income' | 'expense';
}

export default function Dashboard() {
    const router = useRouter();

    // Auth & data state
    const [token, setToken] = useState<string | null>(null);
    const [transactions, setTransactions] = useState<Transaction[]>([]);
    const [categories, setCategories] = useState<Category[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    // Modal/Form state
    const [modalOpen, setModalOpen] = useState(false);
    const [editingTx, setEditingTx] = useState<Transaction | null>(null);
    const [form, setForm] = useState({
        category_id: 0,
        amount: '',
        description: '',
        date: '', // YYYY-MM-DD format
    });
    const [submitting, setSubmitting] = useState(false);

    // --- Effect: Check auth and fetch data ---
    useEffect(() => {
        const storedToken = localStorage.getItem('jwt_token');
        if (!storedToken) {
            router.push('/login');
            return;
        }
        setToken(storedToken);

        async function fetchData() {
            setLoading(true);
            setError(null);
            try {
                // Fetch categories
                const catRes = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/category/list`, {
                    headers: { Authorization: `Bearer ${storedToken}` },
                });
                const catData = await catRes.json();
                if (!catData.success) throw new Error(catData.error || 'Failed to load categories');
                setCategories(catData.categories);

                // Fetch transactions
                const txRes = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/transaction/list`, {
                    headers: { Authorization: `Bearer ${storedToken}` },
                });
                const txData = await txRes.json();
                if (!txData.success) throw new Error(txData.error || 'Failed to load transactions');
                setTransactions(txData.transactions);
            } catch (err: any) {
                setError(err.message);
            } finally {
                setLoading(false);
            }
        }

        fetchData();
    }, [router]);

    // --- Helpers ---
    function openAddModal() {
        setEditingTx(null);
        setForm({
            category_id: categories.length > 0 ? categories[0].id : 0,
            amount: '',
            description: '',
            date: new Date().toISOString().slice(0, 10), // today yyyy-mm-dd
        });
        setModalOpen(true);
        setError(null);
    }

    function openEditModal(tx: Transaction) {
        setEditingTx(tx);
        setForm({
            category_id: tx.category_id,
            amount: tx.amount.toString(),
            description: tx.description,
            date: tx.date.slice(0, 10),
        });
        setModalOpen(true);
        setError(null);
    }

    function closeModal() {
        setModalOpen(false);
        setError(null);
        setSubmitting(false);
    }

    async function handleSubmit(e: React.FormEvent) {
        e.preventDefault();
        if (!token) return;
        setSubmitting(true);
        setError(null);

        // Validate form
        const { category_id, amount, description, date } = form;
        if (!category_id) {
            setError('Please select a category');
            setSubmitting(false);
            return;
        }
        if (!amount || isNaN(Number(amount)) || Number(amount) <= 0) {
            setError('Enter a valid amount greater than zero');
            setSubmitting(false);
            return;
        }
        if (!date) {
            setError('Please select a date');
            setSubmitting(false);
            return;
        }

        try {
            const formData = new FormData();
            formData.append('category_id', String(category_id));
            formData.append('amount', amount);
            formData.append('description', description);
            formData.append('date', date);

            const url = editingTx
                ? `${process.env.NEXT_PUBLIC_API_URL}/transaction/update`
                : `${process.env.NEXT_PUBLIC_API_URL}/transaction/add`;

            if (editingTx) {
                formData.append('id', String(editingTx.id));
            }

            const res = await fetch(url, {
                method: 'POST',
                headers: { Authorization: `Bearer ${token}` },
                body: formData,
            });

            const data = await res.json();
            if (!data.success) {
                throw new Error(data.error || 'Failed to save transaction');
            }

            // Refresh transactions after add/edit
            const txRes = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/transaction/list`, {
                headers: { Authorization: `Bearer ${token}` },
            });
            const txData = await txRes.json();
            if (!txData.success) throw new Error(txData.error || 'Failed to load transactions');
            setTransactions(txData.transactions);

            closeModal();
        } catch (err: any) {
            setError(err.message);
            setSubmitting(false);
        }
    }

    async function handleDelete(id: number) {
        if (!token) return;
        if (!window.confirm('Are you sure you want to delete this transaction?')) return;
        setError(null);

        try {
            const formData = new FormData();
            formData.append('id', String(id));

            const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/transaction/delete`, {
                method: 'POST',
                headers: { Authorization: `Bearer ${token}` },
                body: formData,
            });

            const data = await res.json();
            if (!data.success) throw new Error(data.error || 'Failed to delete transaction');

            setTransactions(transactions.filter(tx => tx.id !== id));
        } catch (err: any) {
            setError(err.message);
        }
    }

    if (loading) return <div className="p-4 text-center">Loading...</div>;

    if (error)
        return (
            <div className="p-4 text-center text-red-600 font-semibold">
                Error: {error}
            </div>
        );

    return (
        <div className="max-w-6xl mx-auto p-6">
            <h1 className="text-3xl font-bold mb-6 text-center">Dashboard</h1>

            <div className="flex justify-end mb-4">
                <button
                    onClick={openAddModal}
                    className="bg-indigo-600 hover:bg-indigo-700 text-white px-4 py-2 rounded font-semibold"
                >
                    + Add Transaction
                </button>
            </div>

            {transactions.length === 0 ? (
                <p className="text-center text-gray-600">No transactions found.</p>
            ) : (
                <table className="w-full border-collapse border border-gray-300 rounded-lg shadow-sm overflow-hidden text-sm">
                    <thead className="bg-gray-100">
                        <tr>
                            <th className="p-3 border border-gray-300 text-left">Date</th>
                            <th className="p-3 border border-gray-300 text-left">Description</th>
                            <th className="p-3 border border-gray-300 text-left">Category</th>
                            <th className="p-3 border border-gray-300 text-right">Amount</th>
                            <th className="p-3 border border-gray-300 text-center">Actions</th>
                        </tr>
                    </thead>
                    <tbody>
                        {transactions.map(tx => {
                            const category = categories.find(c => c.id === tx.category_id);
                            return (
                                <tr
                                    key={tx.id}
                                    className="hover:bg-gray-50 transition-colors duration-150"
                                >
                                    <td className="p-3 border border-gray-300">
                                        {new Date(tx.date).toLocaleDateString()}
                                    </td>
                                    <td className="p-3 border border-gray-300">{tx.description}</td>
                                    <td className="p-3 border border-gray-300">
                                        {category ? category.name : 'Unknown'}
                                    </td>
                                    <td className="p-3 border border-gray-300 text-right">
                                        ${tx.amount.toFixed(2)}
                                    </td>
                                    <td className="p-3 border border-gray-300 text-center space-x-2">
                                        <button
                                            onClick={() => openEditModal(tx)}
                                            className="text-blue-600 hover:underline"
                                            title="Edit"
                                        >
                                            Edit
                                        </button>
                                        <button
                                            onClick={() => handleDelete(tx.id)}
                                            className="text-red-600 hover:underline"
                                            title="Delete"
                                        >
                                            Delete
                                        </button>
                                    </td>
                                </tr>
                            );
                        })}
                    </tbody>
                </table>
            )}

            {/* Modal */}
            {modalOpen && (
                <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
                    <form
                        onSubmit={handleSubmit}
                        className="bg-white rounded-lg shadow-lg p-6 max-w-md w-full space-y-4"
                    >
                        <h2 className="text-xl font-bold">
                            {editingTx ? 'Edit Transaction' : 'Add Transaction'}
                        </h2>

                        <div>
                            <label htmlFor="category" className="block font-medium mb-1">
                                Category
                            </label>
                            <select
                                id="category"
                                value={form.category_id}
                                onChange={e =>
                                    setForm({ ...form, category_id: Number(e.target.value) })
                                }
                                className="w-full border border-gray-300 rounded px-3 py-2"
                            >
                                {categories.map(cat => (
                                    <option key={cat.id} value={cat.id}>
                                        {cat.name} ({cat.type})
                                    </option>
                                ))}
                            </select>
                        </div>

                        <div>
                            <label htmlFor="amount" className="block font-medium mb-1">
                                Amount
                            </label>
                            <input
                                type="number"
                                id="amount"
                                step="0.01"
                                min="0"
                                value={form.amount}
                                onChange={e => setForm({ ...form, amount: e.target.value })}
                                className="w-full border border-gray-300 rounded px-3 py-2"
                                required
                            />
                        </div>

                        <div>
                            <label htmlFor="description" className="block font-medium mb-1">
                                Description
                            </label>
                            <input
                                type="text"
                                id="description"
                                value={form.description}
                                onChange={e => setForm({ ...form, description: e.target.value })}
                                className="w-full border border-gray-300 rounded px-3 py-2"
                            />
                        </div>

                        <div>
                            <label htmlFor="date" className="block font-medium mb-1">
                                Date
                            </label>
                            <input
                                type="date"
                                id="date"
                                value={form.date}
                                onChange={e => setForm({ ...form, date: e.target.value })}
                                className="w-full border border-gray-300 rounded px-3 py-2"
                                required
                            />
                        </div>

                        <div className="flex justify-end space-x-2 mt-4">
                            <button
                                type="button"
                                onClick={closeModal}
                                className="px-4 py-2 rounded border border-gray-300 hover:bg-gray-100"
                                disabled={submitting}
                            >
                                Cancel
                            </button>
                            <button
                                type="submit"
                                disabled={submitting}
                                className="px-4 py-2 rounded bg-indigo-600 text-white font-semibold hover:bg-indigo-700 disabled:opacity-50"
                            >
                                {submitting
                                    ? editingTx ? 'Updating...' : 'Adding...'
                                    : editingTx ? 'Update' : 'Add'}
                            </button>
                        </div>

                        {error && <p className="text-red-600 mt-2">{error}</p>}
                    </form>
                </div>
            )}
        </div>
    );
}
