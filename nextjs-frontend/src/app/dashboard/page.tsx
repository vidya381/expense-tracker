'use client';

import { useEffect, useState, useRef } from 'react';
import { useRouter } from 'next/navigation';
import { FiDollarSign, FiTrendingUp, FiRepeat, FiPlus, FiArrowUp, FiCalendar, FiEdit2, FiTrash2 } from 'react-icons/fi';
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
    category_type: string; // "income" or "expense"
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
    const [refreshTrigger, setRefreshTrigger] = useState(0);

    // Filters, sorting & pagination
    const [page, setPage] = useState(1);
    const [hasMore, setHasMore] = useState(false);
    const [filterCategory, setFilterCategory] = useState('');
    const [sortOrder, setSortOrder] = useState<'date_desc' | 'date_asc'>('date_desc');
    const [allTransactions, setAllTransactions] = useState<Transaction[]>([]); // Store all transactions
    const [loadingMore, setLoadingMore] = useState(false);
    const [loadMoreError, setLoadMoreError] = useState<string | null>(null);
    const [showScrollTop, setShowScrollTop] = useState(false);
    const [jumpToDate, setJumpToDate] = useState('');
    const [deleteConfirm, setDeleteConfirm] = useState<Transaction | null>(null);
    const [deleting, setDeleting] = useState(false);

    const scrollContainerRef = useRef<HTMLDivElement>(null);

    // Check auth token on mount and redirect if needed
    useEffect(() => {
        if (!initialized) return;
        if (!token) {
            router.replace('/login');
        } else {
            setAuthChecked(true);
        }
    }, [token, initialized, router]);

    // Refresh data when navigating back from add transaction page
    useEffect(() => {
        if (authChecked && token) {
            // Check if we should refresh (e.g., after adding a transaction)
            const shouldRefresh = sessionStorage.getItem('refreshDashboard');
            if (shouldRefresh === 'true') {
                sessionStorage.removeItem('refreshDashboard');
                setRefreshTrigger(prev => prev + 1);
            }
        }
    }, [authChecked, token]);

    // Refresh data when page becomes visible or focused (e.g., after adding a transaction)
    useEffect(() => {
        function handleVisibilityChange() {
            if (!document.hidden && authChecked && token) {
                setRefreshTrigger(prev => prev + 1);
            }
        }

        function handleFocus() {
            if (authChecked && token) {
                setRefreshTrigger(prev => prev + 1);
            }
        }

        document.addEventListener('visibilitychange', handleVisibilityChange);
        window.addEventListener('focus', handleFocus);

        return () => {
            document.removeEventListener('visibilitychange', handleVisibilityChange);
            window.removeEventListener('focus', handleFocus);
        };
    }, [authChecked, token]);

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

                // 4. Transactions list with filtering/sorting - Fetch ALL transactions
                const params = new URLSearchParams({
                    month: selectedMonth,
                    sort: sortOrder,
                    limit: '1000', // High limit to get all transactions
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

                // Store all transactions
                setAllTransactions(uniqueTransactions);
                setTransactions(uniqueTransactions);
                setPage(1);
                // No pagination - we fetch all transactions at once
                setHasMore(false);
            } catch (err: any) {
                setError(err.message || 'Unknown error');
            } finally {
                setLoading(false);
            }
        }

        fetchData();
    }, [token, authChecked, selectedMonth, sortOrder, refreshTrigger]);

    // Client-side filtering when category filter changes
    useEffect(() => {
        if (!filterCategory) {
            setTransactions(allTransactions);
        } else {
            const filtered = allTransactions.filter((tx: Transaction) =>
                tx.category.toLowerCase().includes(filterCategory.toLowerCase())
            );
            setTransactions(filtered);
        }
    }, [filterCategory, allTransactions]);

    // Load more handler for pagination
    async function loadMore() {
        if (!token || loadingMore || !hasMore) return;

        setLoadingMore(true);
        setLoadMoreError(null);

        const nextPage = page + 1;
        const params = new URLSearchParams({
            month: selectedMonth,
            category: filterCategory,
            sort: sortOrder,
            page: `${nextPage}`,
            limit: '20', // Explicitly set page size
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
            const existingIds = new Set(transactions.map(tx => tx.id));
            const uniqueNewTransactions = newTransactions.filter(tx => !existingIds.has(tx.id));

            if (uniqueNewTransactions.length > 0) {
                setTransactions((prev) => [...prev, ...uniqueNewTransactions]);
                setPage(nextPage);
            }

            // Set hasMore to false if we got less than 20 new transactions
            setHasMore(uniqueNewTransactions.length === 20);
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

    // Handle scroll in the transaction container
    const handleScroll = (e: React.UIEvent<HTMLDivElement>) => {
        const target = e.target as HTMLDivElement;
        setShowScrollTop(target.scrollTop > 200);
    };

    // Scroll to top of transaction list
    const scrollToTop = () => {
        if (scrollContainerRef.current) {
            scrollContainerRef.current.scrollTo({ top: 0, behavior: 'smooth' });
        }
    };

    // Handle edit transaction
    const handleEdit = (transactionId: number) => {
        router.push(`/transaction/edit/${transactionId}`);
    };

    // Handle delete transaction
    const handleDelete = async (transactionId: number) => {
        if (!token) return;

        setDeleting(true);
        try {
            const formData = new FormData();
            formData.append('id', String(transactionId));

            const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/transaction/delete`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`
                },
                body: formData
            });

            if (!response.ok) {
                throw new Error('Failed to delete transaction');
            }

            // Remove from local state
            setAllTransactions(prev => prev.filter(tx => tx.id !== transactionId));
            setTransactions(prev => prev.filter(tx => tx.id !== transactionId));
            setDeleteConfirm(null);
            setRefreshTrigger(prev => prev + 1); // Refresh summary data
        } catch (err) {
            console.error('Error deleting transaction:', err);
            alert('Failed to delete transaction. Please try again.');
        } finally {
            setDeleting(false);
        }
    };

    // Jump to specific date
    const handleJumpToDate = () => {
        if (!jumpToDate || !scrollContainerRef.current) return;

        const targetDate = jumpToDate; // Keep as string for comparison (YYYY-MM-DD)
        const transactionElements = scrollContainerRef.current.querySelectorAll('[data-transaction-date]');

        // Find the transaction with the exact date or closest date
        let foundElement: HTMLElement | null = null;
        let closestDiff = Infinity;

        for (let i = 0; i < transactionElements.length; i++) {
            const element = transactionElements[i] as HTMLElement;
            const txDateStr = element.dataset.transactionDate || '';

            // Check for exact match first
            if (txDateStr === targetDate) {
                foundElement = element;
                break;
            }

            // Otherwise, find the closest date
            const diff = Math.abs(new Date(txDateStr).getTime() - new Date(targetDate).getTime());
            if (diff < closestDiff) {
                closestDiff = diff;
                foundElement = element;
            }
        }

        if (foundElement) {
            foundElement.scrollIntoView({ behavior: 'smooth', block: 'start' });
            // Clear the date input after jumping
            setTimeout(() => setJumpToDate(''), 500);
        }
    };

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
                        <div className="flex items-center gap-3">
                            <button
                                onClick={() => router.push('/recurring')}
                                className="px-4 py-2 text-sm font-medium text-white bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-700 hover:to-purple-700 rounded-lg transition-all duration-200 shadow-md hover:shadow-lg flex items-center gap-2"
                            >
                                <FiRepeat className="w-4 h-4" />
                                Recurring
                            </button>
                            <button
                                onClick={logout}
                                className="px-4 py-2 text-sm font-medium text-gray-700 hover:text-indigo-600 transition-colors duration-200"
                            >
                                Logout
                            </button>
                        </div>
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
                        (() => {
                            const totalSpending = spendingData.reduce((sum, item) => sum + item.amount, 0);
                            const maxAmount = Math.max(...spendingData.map(item => item.amount));

                            return (
                                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                                    {spendingData.map((item, index) => {
                                        const percentage = totalSpending > 0 ? (item.amount / totalSpending) * 100 : 0;
                                        const barWidth = maxAmount > 0 ? (item.amount / maxAmount) * 100 : 0;

                                        // Color palette for different categories
                                        const colors = [
                                            { gradient: 'from-indigo-500 to-purple-600', bg: 'from-indigo-50 to-purple-50', border: 'border-indigo-200' },
                                            { gradient: 'from-pink-500 to-rose-600', bg: 'from-pink-50 to-rose-50', border: 'border-pink-200' },
                                            { gradient: 'from-blue-500 to-cyan-600', bg: 'from-blue-50 to-cyan-50', border: 'border-blue-200' },
                                            { gradient: 'from-emerald-500 to-teal-600', bg: 'from-emerald-50 to-teal-50', border: 'border-emerald-200' },
                                            { gradient: 'from-amber-500 to-orange-600', bg: 'from-amber-50 to-orange-50', border: 'border-amber-200' },
                                            { gradient: 'from-violet-500 to-purple-600', bg: 'from-violet-50 to-purple-50', border: 'border-violet-200' },
                                        ];
                                        const color = colors[index % colors.length];

                                        return (
                                            <div
                                                key={item.category}
                                                className={`bg-gradient-to-br ${color.bg} p-5 rounded-xl border-2 ${color.border} shadow-md hover:shadow-xl transition-all duration-200 transform hover:-translate-y-1`}
                                            >
                                                <div className="flex items-start justify-between mb-3">
                                                    <div className="flex-1">
                                                        <h3 className="text-sm font-semibold text-gray-700 mb-1 truncate" title={item.category}>
                                                            {item.category}
                                                        </h3>
                                                        <p className="text-2xl font-bold text-gray-900">
                                                            {formatCurrency(item.amount)}
                                                        </p>
                                                    </div>
                                                    <div className={`p-2 rounded-lg bg-gradient-to-br ${color.gradient} shadow-lg`}>
                                                        <svg className="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
                                                        </svg>
                                                    </div>
                                                </div>

                                                {/* Progress bar */}
                                                <div className="mt-4">
                                                    <div className="flex items-center justify-between mb-2">
                                                        <span className="text-xs font-medium text-gray-600">
                                                            {percentage.toFixed(1)}% of total
                                                        </span>
                                                    </div>
                                                    <div className="w-full bg-white/50 rounded-full h-2.5 overflow-hidden shadow-inner">
                                                        <div
                                                            className={`h-full bg-gradient-to-r ${color.gradient} rounded-full transition-all duration-500 ease-out`}
                                                            style={{ width: `${barWidth}%` }}
                                                        />
                                                    </div>
                                                </div>
                                            </div>
                                        );
                                    })}
                                </div>
                            );
                        })()
                    )}
                </section>

                {/* 3. Transaction History */}
                <section className="bg-white/80 backdrop-blur-sm shadow-xl rounded-2xl p-6 border border-white/20">
                    <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between mb-6 gap-4">
                        <div>
                            <h2 className="text-xl font-bold text-gray-900 mb-1">Transaction History</h2>
                            <p className="text-sm text-gray-600">{transactions.length} transaction{transactions.length !== 1 ? 's' : ''} found</p>
                        </div>
                        <div className="flex flex-wrap items-center gap-3 w-full sm:w-auto">
                            {/* Jump to Date */}
                            <div className="relative flex-1 sm:flex-initial sm:w-36">
                                <input
                                    type="date"
                                    value={jumpToDate}
                                    onChange={e => setJumpToDate(e.target.value)}
                                    onKeyDown={e => {
                                        if (e.key === 'Enter' && jumpToDate) {
                                            handleJumpToDate();
                                        }
                                    }}
                                    className="w-full py-2.5 border-2 border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 transition-all duration-200 bg-white text-gray-900 text-sm"
                                    style={{ paddingLeft: '4px', paddingRight: jumpToDate ? '22px' : '4px' }}
                                    placeholder="Jump to date"
                                />
                                {jumpToDate && (
                                    <button
                                        onClick={handleJumpToDate}
                                        className="absolute inset-y-0 right-1 flex items-center text-indigo-600 hover:text-indigo-800 transition-colors"
                                        title="Jump to date (or press Enter)"
                                    >
                                        <div className="p-1 rounded-md hover:bg-indigo-100 transition-all">
                                            <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M13 7l5 5m0 0l-5 5m5-5H6" />
                                            </svg>
                                        </div>
                                    </button>
                                )}
                            </div>

                            {/* Category Filter */}
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
                                    className="w-full sm:w-52 pl-10 pr-4 py-2.5 border-2 border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 transition-all duration-200 bg-white text-gray-900 placeholder:text-gray-400"
                                />
                            </div>

                            {/* Sort Order */}
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
                        <div>
                            {/* Scrollable Container */}
                            <div
                                ref={scrollContainerRef}
                                onScroll={handleScroll}
                                className="max-h-[600px] overflow-y-auto overflow-x-hidden rounded-xl border-2 border-gray-200 bg-gray-50/50"
                                style={{ scrollBehavior: 'smooth' }}
                            >
                                <div className="divide-y divide-gray-200">
                                    {transactions.map((tx: Transaction, index: number) => {
                                        const isExpense = tx.category_type === 'expense';

                                        return (
                                            <div
                                                key={`${tx.id}-${index}-${tx.date}`}
                                                data-transaction-date={tx.date}
                                                className="group bg-white hover:bg-gradient-to-r hover:from-indigo-50 hover:to-purple-50 transition-all duration-150 px-4 py-3"
                                            >
                                                <div className="flex items-center gap-3">
                                                    {/* Compact Date Badge */}
                                                    <div className="flex-shrink-0">
                                                        <div className="w-12 h-12 rounded-lg bg-gradient-to-br from-indigo-500 to-purple-600 flex flex-col items-center justify-center text-white shadow-sm">
                                                            <span className="text-[10px] font-semibold uppercase leading-none">
                                                                {format(parseISO(tx.date), 'MMM')}
                                                            </span>
                                                            <span className="text-lg font-bold leading-none mt-0.5">
                                                                {format(parseISO(tx.date), 'dd')}
                                                            </span>
                                                        </div>
                                                    </div>

                                                    {/* Transaction Info */}
                                                    <div className="flex-1 min-w-0">
                                                        <div className="flex items-center gap-2 mb-1">
                                                            <h4 className="text-sm font-semibold text-gray-900 truncate">
                                                                {tx.description || 'No description'}
                                                            </h4>
                                                        </div>
                                                        <div className="flex items-center gap-2">
                                                            <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-xs font-medium bg-indigo-100 text-indigo-700">
                                                                {tx.category}
                                                            </span>
                                                            <span className="text-xs text-gray-500">
                                                                {format(parseISO(tx.date), 'MMM dd, yyyy')}
                                                            </span>
                                                        </div>
                                                    </div>

                                                    {/* Amount */}
                                                    <div className="flex-shrink-0 text-right">
                                                        <div className={`text-lg font-bold ${isExpense ? 'text-red-600' : 'text-green-600'}`}>
                                                            {isExpense ? '-' : '+'}{formatCurrency(Math.abs(tx.amount))}
                                                        </div>
                                                    </div>

                                                    {/* Action Buttons */}
                                                    <div className="flex-shrink-0 flex items-center gap-1 ml-2 opacity-0 group-hover:opacity-100 transition-opacity duration-200">
                                                        <button
                                                            onClick={() => handleEdit(tx.id)}
                                                            className="p-2 rounded-lg text-indigo-600 hover:bg-indigo-100 transition-all duration-150"
                                                            title="Edit transaction"
                                                        >
                                                            <FiEdit2 className="w-4 h-4" />
                                                        </button>
                                                        <button
                                                            onClick={() => setDeleteConfirm(tx)}
                                                            className="p-2 rounded-lg text-red-600 hover:bg-red-100 transition-all duration-150"
                                                            title="Delete transaction"
                                                        >
                                                            <FiTrash2 className="w-4 h-4" />
                                                        </button>
                                                    </div>
                                                </div>
                                            </div>
                                        );
                                    })}
                                </div>
                            </div>
                        </div>
                    )}

                </section>
            </div>

            {/* Delete Confirmation Modal */}
            {deleteConfirm && (
                <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
                    <div className="bg-white rounded-2xl shadow-2xl max-w-md w-full p-6 transform transition-all">
                        <div className="flex items-center gap-4 mb-4">
                            <div className="p-3 rounded-full bg-red-100">
                                <FiTrash2 className="w-6 h-6 text-red-600" />
                            </div>
                            <div>
                                <h3 className="text-xl font-bold text-gray-900">Delete Transaction</h3>
                                <p className="text-sm text-gray-600">This action cannot be undone</p>
                            </div>
                        </div>

                        {/* Transaction Details */}
                        <div className="mb-6 p-4 bg-gray-50 rounded-xl border border-gray-200">
                            <div className="flex items-start justify-between mb-3">
                                <div className="flex-1">
                                    <div className="flex items-center gap-2 mb-2">
                                        <span className="text-xs font-semibold px-2 py-1 rounded-full bg-gray-200 text-gray-700">
                                            {deleteConfirm.category}
                                        </span>
                                    </div>
                                    <p className="text-sm text-gray-600">
                                        {deleteConfirm.description || 'No description'}
                                    </p>
                                </div>
                                <div className="text-right">
                                    <p className={`text-lg font-bold ${
                                        deleteConfirm.category_type === 'income'
                                            ? 'text-green-600'
                                            : 'text-red-600'
                                    }`}>
                                        {deleteConfirm.category_type === 'income' ? '+' : '-'}${Math.abs(deleteConfirm.amount).toFixed(2)}
                                    </p>
                                </div>
                            </div>
                            <p className="text-xs text-gray-500">
                                {format(parseISO(deleteConfirm.date), 'MMM dd, yyyy')}
                            </p>
                        </div>

                        <p className="text-gray-700 mb-6 text-sm">
                            Are you sure you want to delete this transaction?
                        </p>

                        <div className="flex gap-3 justify-end">
                            <button
                                onClick={() => setDeleteConfirm(null)}
                                disabled={deleting}
                                className="px-4 py-2 border-2 border-gray-300 rounded-xl font-semibold text-gray-700 hover:bg-gray-100 transition-all duration-200 disabled:opacity-50"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={() => handleDelete(deleteConfirm.id)}
                                disabled={deleting}
                                className="px-4 py-2 bg-gradient-to-r from-red-600 to-red-700 text-white rounded-xl font-semibold hover:from-red-700 hover:to-red-800 transition-all duration-200 shadow-lg hover:shadow-xl disabled:opacity-50"
                            >
                                {deleting ? 'Deleting...' : 'Delete'}
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {/* Scroll to Top Button - Fixed to Viewport */}
            {showScrollTop && (
                <button
                    onClick={scrollToTop}
                    className="fixed bottom-24 right-6 p-3 bg-white border-2 border-indigo-200 text-indigo-600 rounded-full shadow-xl hover:bg-indigo-50 hover:border-indigo-300 focus:outline-none focus:ring-4 focus:ring-indigo-500/50 transition-all duration-200 transform hover:scale-110 z-40 animate-fade-in"
                    aria-label="Scroll to top"
                >
                    <FiArrowUp className="w-5 h-5" />
                </button>
            )}

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
