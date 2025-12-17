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
    const { token, initialized, logout } = useAuth();

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
    const [loadingMore, setLoadingMore] = useState(false);
    const [loadMoreError, setLoadMoreError] = useState<string | null>(null);

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
                setSummary(summaryJson || {
                    total_expenses: 0,
                    total_income: 0,
                    recurring_expenses: 0
                });

                // 2. Recurring expenses (current month)
                const recurringRes = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/recurring/list`, {
                    headers: { Authorization: `Bearer ${token}` }
                });
                if (!recurringRes.ok) throw new Error('Failed to fetch recurring expenses');
                const recurringList: any[] = await recurringRes.json() || [];

                const recurringExpenses = (Array.isArray(recurringList) ? recurringList : []).reduce((acc, item) => {
                    if (item && item.recurrence && item.amount && item.category_id && item.amount > 0) {
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
                    Array.isArray(spendingJson) 
                        ? spendingJson.map((item: any) => ({
                            category: item.category || 'Unknown',
                            amount: item.total || 0,
                        }))
                        : []
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
                
                // Ensure no duplicates in initial load
                const uniqueTransactions = Array.isArray(txData) 
                    ? txData.filter((tx: Transaction, index: number, self: Transaction[]) => 
                        index === self.findIndex((t: Transaction) => t.id === tx.id)
                    )
                    : [];

                setTransactions(uniqueTransactions);
                setPage(1);
                setHasMore(uniqueTransactions.length >= 20); // adjust page size if your API differs
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
        if (!token || loadingMore) return;
        
        setLoadingMore(true);
        setLoadMoreError(null);
        
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
            
            if (!res.ok) {
                throw new Error('Failed to load more transactions');
            }
            
            const moreTx = await res.json();
            const newTransactions = Array.isArray(moreTx) ? moreTx : [];
            
            // Filter out duplicates by ID before adding
            setTransactions((prev) => {
                const existingIds = new Set(prev.map(tx => tx.id));
                const uniqueNewTransactions = newTransactions.filter(tx => !existingIds.has(tx.id));
                return [...prev, ...uniqueNewTransactions];
            });
            
            setPage(nextPage);
            setHasMore(newTransactions.length >= 20);
        } catch (err: any) {
            // Use local error state instead of global error to avoid breaking the page
            setLoadMoreError(err.message || 'Failed to load more transactions. Please try again.');
            console.error('Error loading more transactions:', err);
        } finally {
            setLoadingMore(false);
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
    if (!authChecked || loading) {
        return (
            <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-indigo-50 via-white to-purple-50">
                <div className="text-center">
                    <div className="inline-block animate-spin rounded-full h-12 w-12 border-4 border-indigo-600 border-t-transparent mb-4"></div>
                    <p className="text-gray-600 font-medium">Loading your dashboard...</p>
                </div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-indigo-50 via-white to-purple-50 px-4">
                <div className="bg-white/80 backdrop-blur-sm shadow-2xl rounded-2xl p-8 max-w-md w-full border border-white/20">
                    <div className="text-center">
                        <div className="inline-flex items-center justify-center w-16 h-16 bg-red-100 rounded-full mb-4">
                            <svg className="w-8 h-8 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                            </svg>
                        </div>
                        <h2 className="text-2xl font-bold text-gray-900 mb-2">Error</h2>
                        <p className="text-red-600 mb-6">{error}</p>
                        <button
                            onClick={() => window.location.reload()}
                            className="px-6 py-2.5 bg-gradient-to-r from-indigo-600 to-purple-600 text-white rounded-xl font-semibold hover:from-indigo-700 hover:to-purple-700 transition-all duration-200 shadow-lg hover:shadow-xl"
                        >
                            Try Again
                        </button>
                    </div>
                </div>
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-gradient-to-br from-indigo-50 via-white to-purple-50">
            {/* Header */}
            <header className="bg-white/80 backdrop-blur-sm border-b border-gray-200/50 shadow-sm sticky top-0 z-40">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
                    <div className="flex items-center justify-between">
                        <div className="flex items-center space-x-3">
                            <div className="w-10 h-10 bg-gradient-to-br from-indigo-600 to-purple-600 rounded-xl flex items-center justify-center shadow-lg">
                                <FiDollarSign className="w-6 h-6 text-white" />
                            </div>
                            <h1 className="text-2xl font-bold bg-gradient-to-r from-indigo-600 to-purple-600 bg-clip-text text-transparent">
                                Expense Tracker
                            </h1>
                        </div>
                        <button
                            onClick={logout}
                            className="px-4 py-2 text-sm font-medium text-gray-700 hover:text-indigo-600 transition-colors duration-200"
                        >
                            Logout
                        </button>
                    </div>
                </div>
            </header>

            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
                {/* 1. Top Summary Cards */}
                <section className="grid grid-cols-1 md:grid-cols-3 gap-6">
                    <Card 
                        icon={<FiTrendingUp size={32} />} 
                        label="Total Expenses" 
                        value={formatCurrency(summary?.total_expenses || 0)}
                        gradient="from-red-500 to-pink-500"
                        bgGradient="from-red-50 to-pink-50"
                    />
                    <Card 
                        icon={<FiDollarSign size={32} />} 
                        label="Total Income" 
                        value={formatCurrency(summary?.total_income || 0)}
                        gradient="from-green-500 to-emerald-500"
                        bgGradient="from-green-50 to-emerald-50"
                    />
                    <Card 
                        icon={<FiRepeat size={32} />} 
                        label="Recurring Expenses" 
                        value={formatCurrency(summary?.recurring_expenses || 0)}
                        gradient="from-orange-500 to-amber-500"
                        bgGradient="from-orange-50 to-amber-50"
                    />
                </section>

                {/* 2. Spending Breakdown Chart */}
                <section className="bg-white/80 backdrop-blur-sm shadow-xl rounded-2xl p-6 border border-white/20">
                    <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between mb-6 gap-4">
                        <div>
                            <h2 className="text-xl font-bold text-gray-900 mb-1">Spending Breakdown</h2>
                            <p className="text-sm text-gray-600">{format(parseISO(selectedMonth + '-01'), 'MMMM yyyy')}</p>
                        </div>
                        <div className="flex items-center space-x-2 bg-gray-50 rounded-xl p-1">
                            <button 
                                onClick={prevMonth} 
                                className="px-4 py-2 rounded-lg hover:bg-white transition-colors duration-200 text-gray-700 font-medium"
                                aria-label="Previous month"
                            >
                                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
                                </svg>
                            </button>
                            <span className="px-4 py-2 text-sm font-medium text-gray-700 min-w-[120px] text-center">
                                {format(parseISO(selectedMonth + '-01'), 'MMM yyyy')}
                            </span>
                            <button 
                                onClick={nextMonth} 
                                className="px-4 py-2 rounded-lg hover:bg-white transition-colors duration-200 text-gray-700 font-medium disabled:opacity-50 disabled:cursor-not-allowed"
                                aria-label="Next month"
                                disabled={subMonths(parseISO(selectedMonth + '-01'), -1) > new Date()}
                            >
                                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                                </svg>
                            </button>
                        </div>
                    </div>
                    {spendingData.length === 0 ? (
                        <div className="h-[300px] flex items-center justify-center text-gray-500">
                            <div className="text-center">
                                <svg className="w-16 h-16 mx-auto mb-4 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                                </svg>
                                <p>No spending data available for this month</p>
                            </div>
                        </div>
                    ) : (
                        <ResponsiveContainer width="100%" height={300}>
                            <BarChart data={spendingData}>
                                <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                                <XAxis 
                                    dataKey="category" 
                                    tick={{ fill: '#6b7280', fontSize: 12 }}
                                    angle={-45}
                                    textAnchor="end"
                                    height={80}
                                />
                                <YAxis 
                                    tick={{ fill: '#6b7280', fontSize: 12 }}
                                    tickFormatter={(value) => `$${value}`}
                                />
                                <Tooltip 
                                    formatter={(value: number) => formatCurrency(value)}
                                    contentStyle={{ 
                                        backgroundColor: 'white', 
                                        border: '1px solid #e5e7eb', 
                                        borderRadius: '8px',
                                        boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.1)'
                                    }}
                                />
                                <Bar 
                                    dataKey="amount" 
                                    fill="url(#colorGradient)"
                                    radius={[8, 8, 0, 0]}
                                />
                                <defs>
                                    <linearGradient id="colorGradient" x1="0" y1="0" x2="0" y2="1">
                                        <stop offset="0%" stopColor="#4f46e5" stopOpacity={1}/>
                                        <stop offset="100%" stopColor="#7c3aed" stopOpacity={1}/>
                                    </linearGradient>
                                </defs>
                            </BarChart>
                        </ResponsiveContainer>
                    )}
                </section>

                {/* 3. Transaction History */}
                <section className="bg-white/80 backdrop-blur-sm shadow-xl rounded-2xl p-6 border border-white/20">
                    <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between mb-6 gap-4">
                        <h2 className="text-xl font-bold text-gray-900">Transaction History</h2>
                        <div className="flex flex-wrap items-center gap-3 w-full sm:w-auto">
                            <div className="relative flex-1 sm:flex-initial">
                                <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                                    <svg className="h-5 w-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                                    </svg>
                                </div>
                                <input
                                    type="text"
                                    placeholder="Filter by category"
                                    value={filterCategory}
                                    onChange={e => setFilterCategory(e.target.value)}
                                    className="w-full sm:w-64 pl-10 pr-4 py-2.5 border-2 border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 transition-all duration-200 bg-white text-gray-900 placeholder:text-gray-400"
                                />
                            </div>
                            <select
                                value={sortOrder}
                                onChange={e => setSortOrder(e.target.value as 'date_desc' | 'date_asc')}
                                className="px-4 py-2.5 border-2 border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 transition-all duration-200 bg-white text-gray-900 font-medium"
                            >
                                <option value="date_desc">Newest First</option>
                                <option value="date_asc">Oldest First</option>
                            </select>
                        </div>
                    </div>

                    {transactions.length === 0 ? (
                        <div className="text-center py-12">
                            <svg className="w-20 h-20 mx-auto mb-4 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                            </svg>
                            <p className="text-gray-500 font-medium">No transactions found</p>
                            <p className="text-sm text-gray-400 mt-1">Try adjusting your filters or add a new transaction</p>
                        </div>
                    ) : (
                        <div className="overflow-x-auto rounded-xl border border-gray-200">
                            <table className="min-w-full divide-y divide-gray-200">
                                <thead className="bg-gradient-to-r from-gray-50 to-gray-100/50">
                                    <tr>
                                        <th className="px-6 py-4 text-left text-xs font-semibold text-gray-700 uppercase tracking-wider">Date</th>
                                        <th className="px-6 py-4 text-left text-xs font-semibold text-gray-700 uppercase tracking-wider">Description</th>
                                        <th className="px-6 py-4 text-left text-xs font-semibold text-gray-700 uppercase tracking-wider">Category</th>
                                        <th className="px-6 py-4 text-right text-xs font-semibold text-gray-700 uppercase tracking-wider">Amount</th>
                                    </tr>
                                </thead>
                                <tbody className="bg-white divide-y divide-gray-200">
                                    {transactions.map((tx: Transaction, index: number) => (
                                        <tr key={`${tx.id}-${index}-${tx.date}`} className="hover:bg-indigo-50/50 transition-colors duration-150">
                                            <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                                                {format(parseISO(tx.date), 'MMM dd, yyyy')}
                                            </td>
                                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-700">{tx.description || '—'}</td>
                                            <td className="px-6 py-4 whitespace-nowrap">
                                                <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-indigo-100 text-indigo-800">
                                                    {tx.category}
                                                </span>
                                            </td>
                                            <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-semibold text-gray-900">
                                                {formatCurrency(tx.amount)}
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    )}

                    {loadMoreError && (
                        <div className="mt-4 p-4 bg-red-50 border-l-4 border-red-500 rounded-lg">
                            <div className="flex items-center">
                                <svg className="w-5 h-5 text-red-600 mr-2" fill="currentColor" viewBox="0 0 20 20">
                                    <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                                </svg>
                                <span className="text-sm text-red-700">{loadMoreError}</span>
                                <button
                                    onClick={() => setLoadMoreError(null)}
                                    className="ml-auto text-red-600 hover:text-red-800"
                                    aria-label="Dismiss error"
                                >
                                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                                    </svg>
                                </button>
                            </div>
                        </div>
                    )}

                    {hasMore && (
                        <div className="text-center mt-6">
                            <button
                                onClick={loadMore}
                                disabled={loadingMore}
                                className={`px-6 py-2.5 bg-gradient-to-r from-indigo-600 to-purple-600 text-white rounded-xl font-semibold transition-all duration-200 shadow-lg hover:shadow-xl disabled:opacity-50 disabled:cursor-not-allowed ${
                                    !loadingMore ? 'hover:from-indigo-700 hover:to-purple-700' : ''
                                }`}
                            >
                                {loadingMore ? (
                                    <span className="flex items-center justify-center">
                                        <svg className="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                                            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                                            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                                        </svg>
                                        Loading...
                                    </span>
                                ) : (
                                    'Load More'
                                )}
                            </button>
                        </div>
                    )}
                </section>
            </div>

            {/* Floating Add Transaction Button */}
            <button
                onClick={() => router.push('/transaction/add')}
                aria-label="Add transaction"
                className="fixed bottom-6 right-6 p-5 rounded-full shadow-2xl bg-gradient-to-r from-indigo-600 to-purple-600 text-white hover:from-indigo-700 hover:to-purple-700 focus:outline-none focus:ring-4 focus:ring-indigo-500/50 transition-all duration-200 transform hover:scale-110 active:scale-95 z-50"
            >
                <FiPlus size={28} />
            </button>
        </div>
    );
}

function Card({ icon, label, value, gradient, bgGradient }: { 
    icon: React.ReactNode; 
    label: string; 
    value: string;
    gradient: string;
    bgGradient: string;
}) {
    return (
        <div className={`bg-gradient-to-br ${bgGradient} p-6 rounded-2xl shadow-lg border border-white/50 backdrop-blur-sm hover:shadow-xl transition-all duration-200 transform hover:-translate-y-1`}>
            <div className="flex items-center justify-between mb-4">
                <div className={`p-3 rounded-xl bg-gradient-to-br ${gradient} shadow-lg`}>
                    <div className="text-white">{icon}</div>
                </div>
            </div>
            <div>
                <p className="text-3xl font-bold text-gray-900 mb-1">{value}</p>
                <p className="text-sm font-medium text-gray-600">{label}</p>
            </div>
        </div>
    );
}
