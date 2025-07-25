'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import TransactionForm from '../../components/TransactionForm';

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

export default function DashboardPage() {
    const router = useRouter();
    const [token, setToken] = useState<string | null>(null);
    const [transactions, setTransactions] = useState<Transaction[]>([]);
    const [categories, setCategories] = useState<Category[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const [modalOpen, setModalOpen] = useState(false);
    const [editingTx, setEditingTx] = useState<Transaction | null>(null);

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

    function openAddModal() {
        setEditingTx(null);
        setModalOpen(true);
    }

    function openEditModal(tx: Transaction) {
        setEditingTx(tx);
        setModalOpen(true);
    }

    function closeModal() {
        setModalOpen(false);
    }

    async function refreshTransactions() {
        if (!token) return;
        try {
            const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/transaction/list`, {
                headers: { Authorization: `Bearer ${token}` },
            });
            const data = await res.json();
            if (data.success) {
                setTransactions(data.transactions);
            } else {
                setError(data.error || 'Failed to reload transactions');
            }
        } catch (err: any) {
            setError(err.message);
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

            {modalOpen && (
                <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50 p-4">
                    <div className="bg-white rounded-lg shadow-lg p-6 max-w-md w-full overflow-auto max-h-[90vh]">
                        <TransactionForm
                            token={token!}
                            categories={categories}
                            setCategories={setCategories}
                            onSuccess={() => {
                                refreshTransactions();
                                closeModal();
                            }}
                            onCancel={closeModal}
                            initialValues={
                                editingTx
                                    ? {
                                        id: editingTx.id,
                                        category_id: editingTx.category_id,
                                        category_name:
                                            categories.find(c => c.id === editingTx.category_id)?.name ||
                                            '',
                                        amount: editingTx.amount,
                                        description: editingTx.description,
                                        date: editingTx.date.slice(0, 10),
                                    }
                                    : undefined
                            }
                        />
                    </div>
                </div>
            )}
        </div>
    );
}
