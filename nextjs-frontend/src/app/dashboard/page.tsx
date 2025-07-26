'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import {
    BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid,
} from 'recharts';
import { FiDollarSign, FiTrendingUp, FiRepeat, FiPlus } from 'react-icons/fi';
import { useAuth } from '../../context/AuthContext';
import { format, parseISO, subMonths } from 'date-fns';

interface SummaryData {
    total_expenses: number;
    total_income: number;
    recurring_expenses: number;
}

interface SpendingCategory {
    category: string;
    amount: number;
}

interface Transaction {
    id: number;
    date: string;
    amount: number;
    category: string;
    description: string;
}

export default function Dashboard() {
    const router = useRouter();
    const { token, initialized } = useAuth();

    // Track auth check is complete
    const [authChecked, setAuthChecked] = useState(false);

    // States
    const [summary, setSummary] = useState<SummaryData | null>(null);
    const [selectedMonth, setSelectedMonth] = useState<string>(format(new Date(), 'yyyy-MM'));
    const [spendingData, setSpendingData] = useState<SpendingCategory[]>([]);
    const [transactions, setTransactions] = useState<Transaction[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    // Filters, sorting & pagination
    const [page, setPage] = useState(1);
    const [hasMore, setHasMore] = useState(false);
    const [filterCategory, setFilterCategory] = useState('');
    const [sortOrder, setSortOrder] = useState<'date_desc' | 'date_asc'>('date_desc');

    // Check auth token on mount and redirect if needed
    useEffect(() => {
        if (!initialized) return;
        if (!token) {
            router.replace('/login');
        } else {
            setAuthChecked(true);
        }
    }, [token, initialized, router]);

    // Fetch data whenever token or filters/month changes
    useEffect(() => {
        if (!token || !authChecked) return;

        setLoading(true);
        setError(null);

        async function fetchData() {
            try {
                // 1. Summary totals
                const summaryRes = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/summary/totals`, {
                    headers: { Authorization: `Bearer ${token}` }
                });
                if (!summaryRes.ok) throw new Error('Failed to fetch summary');
                const summaryJson = await summaryRes.json();
                setSummary(summaryJson);

                // 2. Recurring expenses (current month)
                const recurringRes = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/recurring/list`, {
                    headers: { Authorization: `Bearer ${token}` }
                });
                if (!recurringRes.ok) throw new Error('Failed to fetch recurring expenses');
                const recurringList: any[] = await recurringRes.json();

                const recurringExpenses = recurringList.reduce((acc, item) => {
                    if (item.recurrence && item.amount && item.category_id && item.amount > 0) {
                        acc += item.amount;
                    }
                    return acc;
                }, 0);
                setSummary(prev => prev ? { ...prev, recurring_expenses: recurringExpenses } : null);

                // 3. Spending breakdown by category for month
                const spendingRes = await fetch(
                    `${process.env.NEXT_PUBLIC_API_URL}/summary/category/monthly?year=${selectedMonth.slice(0, 4)}&month=${parseInt(selectedMonth.slice(5, 7))}`,
                    { headers: { Authorization: `Bearer ${token}` } }
                );
                if (!spendingRes.ok) throw new Error('Failed to fetch spending breakdown');
                const spendingJson = await spendingRes.json();
                setSpendingData(
                    spendingJson.map((item: any) => ({
                        category: item.category,
                        amount: item.total,
                    }))
                );

                // 4. Transactions list with filtering/sorting
                const params = new URLSearchParams({
                    month: selectedMonth,
                    category: filterCategory,
                    sort: sortOrder,
                    page: '1',
                });

                const txRes = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/transactions/search?${params.toString()}`, {
                    headers: { Authorization: `Bearer ${token}` }
                });
                if (!txRes.ok) throw new Error('Failed to fetch transactions');
                const txData = await txRes.json();

                setTransactions(txData);
                setPage(1);
                setHasMore(txData.length >= 20); // adjust page size if your API differs
            } catch (err: any) {
                setError(err.message || 'Unknown error');
            } finally {
                setLoading(false);
            }
        }

        fetchData();
    }, [token, authChecked, selectedMonth, filterCategory, sortOrder]);

    // Load more handler for pagination
    async function loadMore() {
        if (!token) return;
        const nextPage = page + 1;
        const params = new URLSearchParams({
            month: selectedMonth,
            category: filterCategory,
            sort: sortOrder,
            page: `${nextPage}`,
        });
        try {
            const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/transactions/search?${params.toString()}`, {
                headers: { Authorization: `Bearer ${token}` },
            });
            if (!res.ok) throw new Error('Failed to load more transactions');
            const moreTx = await res.json();
            setTransactions((prev) => [...prev, ...moreTx]);
            setPage(nextPage);
            setHasMore(moreTx.length >= 20);
        } catch (err: any) {
            setError(err.message);
        }
    }

    // Format currencies nicely
    function formatCurrency(value: number) {
        return value.toLocaleString(undefined, { style: 'currency', currency: 'USD' });
    }

    // Handlers for month slider (prev/next month)
    function prevMonth() {
        const prev = subMonths(parseISO(selectedMonth + '-01'), 1);
        setSelectedMonth(format(prev, 'yyyy-MM'));
    }
    function nextMonth() {
        const next = subMonths(parseISO(selectedMonth + '-01'), -1);
        if (next > new Date()) return; // disallow future months
        setSelectedMonth(format(next, 'yyyy-MM'));
    }

    // Show loading screen until auth check done or data loading
    if (!authChecked || loading) return <div className="p-4 text-center">Loading...</div>;

    if (error) return <div className="p-4 text-center text-red-600 font-semibold">Error: {error}</div>;

    return (
        <div className="max-w-7xl mx-auto p-4 space-y-8 min-h-screen">

            {/* 1. Top Summary Cards */}
            <section className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <Card icon={<FiTrendingUp size={28} className="text-indigo-600" />} label="Total Expenses" value={formatCurrency(summary?.total_expenses || 0)} />
                <Card icon={<FiDollarSign size={28} className="text-green-600" />} label="Total Income" value={formatCurrency(summary?.total_income || 0)} />
                <Card icon={<FiRepeat size={28} className="text-red-600" />} label="Recurring Expenses" value={formatCurrency(summary?.recurring_expenses || 0)} />
            </section>

            {/* 2. Spending Breakdown Chart */}
            <section className="bg-white p-6 rounded shadow">
                <div className="flex items-center justify-between mb-4">
                    <h2 className="font-semibold text-lg">Spending Breakdown - {format(parseISO(selectedMonth + '-01'), 'MMMM yyyy')}</h2>
                    <div className="flex space-x-2">
                        <button onClick={prevMonth} className="border rounded px-2 py-1 hover:bg-gray-100">{'<'}</button>
                        <button onClick={nextMonth} className="border rounded px-2 py-1 hover:bg-gray-100">{'>'}</button>
                    </div>
                </div>
                <ResponsiveContainer width="100%" height={300}>
                    <BarChart data={spendingData}>
                        <CartesianGrid strokeDasharray="3 3" />
                        <XAxis dataKey="category" />
                        <YAxis />
                        <Tooltip formatter={(value: number) => `$${value}`} />
                        <Bar dataKey="amount" fill="#4f46e5" />
                    </BarChart>
                </ResponsiveContainer>
            </section>

            {/* 3. Transaction History */}
            <section>
                <div className="flex flex-wrap items-center justify-between mb-4 gap-4">
                    <div className="flex gap-2">
                        <input
                            type="text"
                            placeholder="Filter by category"
                            value={filterCategory}
                            onChange={e => setFilterCategory(e.target.value)}
                            className="border rounded px-3 py-1"
                        />
                    </div>
                    <select
                        value={sortOrder}
                        onChange={e => setSortOrder(e.target.value as 'date_desc' | 'date_asc')}
                        className="border rounded px-3 py-1"
                    >
                        <option value="date_desc">Newest First</option>
                        <option value="date_asc">Oldest First</option>
                    </select>
                </div>

                {transactions.length === 0 ? (
                    <p className="text-center text-gray-500 py-6">No transactions found.</p>
                ) : (
                    <div className="overflow-x-auto rounded shadow">
                        <table className="min-w-full bg-white divide-y divide-gray-200">
                            <thead className="bg-gray-50">
                                <tr>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Date</th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Description</th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Category</th>
                                    <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Amount</th>
                                </tr>
                            </thead>
                            <tbody className="divide-y divide-gray-200">
                                {transactions.map((tx: Transaction) => (
                                    <tr key={tx.id} className="hover:bg-gray-100">
                                        <td className="px-6 py-4 whitespace-nowrap">{format(parseISO(tx.date), 'yyyy-MM-dd')}</td>
                                        <td className="px-6 py-4 whitespace-nowrap">{tx.description}</td>
                                        <td className="px-6 py-4 whitespace-nowrap">{tx.category}</td>
                                        <td className="px-6 py-4 whitespace-nowrap text-right">${tx.amount.toFixed(2)}</td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                )}

                {hasMore && (
                    <div className="text-center mt-4">
                        <button
                            onClick={loadMore}
                            className="px-4 py-2 rounded bg-indigo-600 text-white hover:bg-indigo-700"
                        >
                            Load More
                        </button>
                    </div>
                )}
            </section>

            {/* Floating Add Transaction Button */}
            <button
                onClick={() => router.push('/transaction/add')} // Update this if you switch to modal add
                aria-label="Add transaction"
                className="fixed bottom-6 right-6 p-4 rounded-full shadow-lg bg-indigo-600 text-white hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
                <FiPlus size={24} />
            </button>
        </div>
    );
}

function Card({ icon, label, value }: { icon: React.ReactNode; label: string; value: string }) {
    return (
        <div className="bg-white p-6 rounded shadow flex items-center space-x-4">
            <div className="text-indigo-600">{icon}</div>
            <div>
                <p className="text-2xl font-bold">{value}</p>
                <p className="text-gray-500">{label}</p>
            </div>
        </div>
    );
}
