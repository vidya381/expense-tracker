'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import {
    fetchTransactions,
    addTransaction,
    updateTransaction,
    deleteTransaction
} from '../../lib/api';

interface Transaction {
    id: number;
    category_id: number;
    amount: number;
    description: string;
    date: string;
}

function defaultTx(): Omit<Transaction, 'id'> {
    return { category_id: 1, amount: 0, description: '', date: '' };
}

export default function Dashboard() {
    const router = useRouter();
    const [token, setToken] = useState<string | null>(null);

    const [transactions, setTransactions] = useState<Transaction[]>([]);
    const [error, setError] = useState<string | null>(null);
    const [loading, setLoading] = useState(true);

    // Add/Edit modal state
    const [showModal, setShowModal] = useState(false);
    const [modalMode, setModalMode] = useState<'add' | 'edit'>('add');
    const [form, setForm] = useState<any>(defaultTx());
    const [editingId, setEditingId] = useState<number | null>(null);

    useEffect(() => {
        const t = localStorage.getItem('jwt_token');
        if (!t) {
            router.push('/login');
            return;
        }
        setToken(t);
        fetch(`${process.env.NEXT_PUBLIC_API_URL}/transaction/list`, {
            headers: { Authorization: `Bearer ${t}` }
        })
            .then(res => res.json())
            .then(data => {
                if (data.success) setTransactions(data.transactions);
                else setError(data.error || "Failed to load transactions.");
            })
            .catch(e => setError(e.message))
            .finally(() => setLoading(false));
    }, [router]);


    // Modal open handlers
    function startAdd() {
        setModalMode('add');
        setForm(defaultTx());
        setEditingId(null);
        setShowModal(true);
        setError(null);
    }
    function startEdit(tx: Transaction) {
        setModalMode('edit');
        setForm({ ...tx });
        setEditingId(tx.id);
        setShowModal(true);
        setError(null);
    }

    async function handleDelete(id: number) {
        if (!token) return;
        if (!window.confirm("Are you sure you want to delete this transaction?")) return;
        try {
            await deleteTransaction(token, id);
            setTransactions((prev) => prev.filter((tx) => tx.id !== id));
        } catch (e: any) {
            setError(e.message);
        }
    }

    async function handleSubmit(e: React.FormEvent) {
        e.preventDefault();
        if (!token) return;
        setError(null);

        try {
            if (modalMode === 'add') {
                await addTransaction(token, form);
            } else if (modalMode === 'edit' && editingId != null) {
                await updateTransaction(token, { ...form, id: editingId });
            }
            // Refresh list:
            const txs = await fetchTransactions(token);
            setTransactions(txs);
            setShowModal(false);
        } catch (e: any) {
            setError(e.message);
        }
    }

    if (loading) return <div className="p-4 text-center">Loading transactions...</div>;

    return (
        <div className="max-w-5xl mx-auto p-6">
            <h1 className="text-2xl font-bold mb-6 text-center">Dashboard</h1>
            {error && <div className="bg-red-100 border border-red-400 text-red-800 rounded p-2 mb-4">{error}</div>}
            <div className="flex justify-end mb-4">
                <button
                    onClick={startAdd}
                    className="bg-indigo-600 text-white px-4 py-2 rounded font-semibold hover:bg-indigo-700 shadow"
                >
                    + Add Transaction
                </button>
            </div>
            <table className="w-full border-collapse border border-gray-300 shadow-sm rounded-lg overflow-hidden text-sm">
                <thead>
                    <tr className="bg-gray-100">
                        <th className="p-2 border border-gray-300">Date</th>
                        <th className="p-2 border border-gray-300">Description</th>
                        <th className="p-2 border border-gray-300">Amount</th>
                        <th className="p-2 border border-gray-300">Actions</th>
                    </tr>
                </thead>
                <tbody>
                    {transactions.length === 0 && (
                        <tr>
                            <td colSpan={4} className="p-4 text-center text-gray-500">
                                No transactions found.
                            </td>
                        </tr>
                    )}
                    {transactions.map((tx) => (
                        <tr key={tx.id} className="hover:bg-gray-50 transition">
                            <td className="p-2 border">{new Date(tx.date).toLocaleDateString()}</td>
                            <td className="p-2 border">{tx.description}</td>
                            <td className="p-2 border text-right">${tx.amount?.toFixed(2)}</td>
                            <td className="p-2 border text-center">
                                <button
                                    onClick={() => startEdit(tx)}
                                    className="text-blue-600 font-semibold hover:underline mr-4"
                                >
                                    Edit
                                </button>
                                <button
                                    onClick={() => handleDelete(tx.id)}
                                    className="text-red-600 font-semibold hover:underline"
                                >
                                    Delete
                                </button>
                            </td>
                        </tr>
                    ))}
                </tbody>
            </table>
            {/* Modal for add/edit transaction */}
            {showModal && (
                <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50">
                    <form
                        className="bg-white p-8 rounded-lg shadow max-w-xs w-full space-y-4"
                        onSubmit={handleSubmit}
                    >
                        <h2 className="font-bold mb-4">
                            {modalMode === 'add' ? 'Add Transaction' : 'Edit Transaction'}
                        </h2>
                        <div>
                            <label className="block text-sm mb-1">Category ID</label>
                            <input
                                type="number"
                                value={form.category_id}
                                min={1}
                                className="px-3 py-2 border rounded w-full"
                                onChange={e => setForm({ ...form, category_id: Number(e.target.value) })}
                            />
                        </div>
                        <div>
                            <label className="block text-sm mb-1">Amount</label>
                            <input
                                type="number"
                                value={form.amount}
                                min={0.01}
                                step={0.01}
                                className="px-3 py-2 border rounded w-full"
                                onChange={e => setForm({ ...form, amount: Number(e.target.value) })}
                            />
                        </div>
                        <div>
                            <label className="block text-sm mb-1">Description</label>
                            <input
                                type="text"
                                value={form.description}
                                className="px-3 py-2 border rounded w-full"
                                onChange={e => setForm({ ...form, description: e.target.value })}
                            />
                        </div>
                        <div>
                            <label className="block text-sm mb-1">Date</label>
                            <input
                                type="date"
                                value={form.date}
                                className="px-3 py-2 border rounded w-full"
                                onChange={e => setForm({ ...form, date: e.target.value })}
                            />
                        </div>
                        <div className="flex justify-between gap-2 mt-6">
                            <button
                                type="submit"
                                className="bg-indigo-600 text-white px-4 py-2 rounded hover:bg-indigo-700"
                            >
                                {modalMode === 'add' ? 'Add' : 'Update'}
                            </button>
                            <button
                                type="button"
                                className="bg-gray-300 text-gray-800 px-4 py-2 rounded hover:bg-gray-400"
                                onClick={() => setShowModal(false)}
                            >
                                Cancel
                            </button>
                        </div>
                        {!!error && <div className="text-red-500 text-sm mt-2">{error}</div>}
                    </form>
                </div>
            )}
        </div>
    );
}
