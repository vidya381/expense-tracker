'use client';

import { useState, useEffect, useRef } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '../../context/AuthContext';
import { FiRepeat, FiEdit2, FiTrash2, FiPlus, FiArrowLeft, FiCalendar, FiX } from 'react-icons/fi';
import { format, parseISO } from 'date-fns';

// Calendar date helpers - treat dates as pure calendar days without timezone conversion
const MONTH_NAMES_SHORT = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

function formatCalendarDate(dateStr: string): string {
    // Extract just the date part if it's an ISO timestamp (e.g., "2025-12-20T00:00:00Z")
    const datePart = dateStr.split('T')[0];
    const [year, month, day] = datePart.split('-').map(Number);
    return `${MONTH_NAMES_SHORT[month - 1]} ${day}, ${year}`;
}

interface Category {
    id: number;
    name: string;
    type: 'income' | 'expense';
}

interface RecurringTransaction {
    id: number;
    user_id: number;
    category_id: number;
    amount: number;
    description: string;
    start_date: string;
    recurrence: string;
    last_occurrence: string | null;
    created_at: string;
}

export default function RecurringTransactionsPage() {
    const router = useRouter();
    const { token, initialized } = useAuth();
    const [authChecked, setAuthChecked] = useState(false);

    const [recurring, setRecurring] = useState<RecurringTransaction[]>([]);
    const [categories, setCategories] = useState<Category[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const [showModal, setShowModal] = useState(false);
    const [editingId, setEditingId] = useState<number | null>(null);
    const [deleteConfirm, setDeleteConfirm] = useState<RecurringTransaction | null>(null);
    const [deleting, setDeleting] = useState(false);

    // Form fields
    const [categoryInput, setCategoryInput] = useState('');
    const [categoryId, setCategoryId] = useState<number | null>(null);
    const [filteredCategories, setFilteredCategories] = useState<Category[]>([]);
    const [showDropdown, setShowDropdown] = useState(false);
    const [showCategoryTypeInput, setShowCategoryTypeInput] = useState(false);
    const [categoryType, setCategoryType] = useState<'income' | 'expense'>('expense');
    const [amount, setAmount] = useState('');
    const [description, setDescription] = useState('');
    const [startDate, setStartDate] = useState(new Date().toISOString().slice(0, 10));
    const [recurrence, setRecurrence] = useState<'daily' | 'weekly' | 'monthly' | 'yearly'>('monthly');
    const [submitting, setSubmitting] = useState(false);

    const categoryInputRef = useRef<HTMLInputElement>(null);
    const dropdownRef = useRef<HTMLDivElement>(null);

    // Check auth
    useEffect(() => {
        if (!initialized) return;
        if (!token) {
            router.replace('/login');
        } else {
            setAuthChecked(true);
        }
    }, [token, initialized, router]);

    // Fetch data
    useEffect(() => {
        if (!token || !authChecked) return;

        async function fetchData() {
            try {
                // Fetch categories
                const catRes = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/category/list`, {
                    headers: { Authorization: `Bearer ${token}` },
                });
                const catData = await catRes.json();
                if (catData.success) {
                    setCategories(catData.categories);
                }

                // Fetch recurring transactions
                const recRes = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/recurring/list`, {
                    headers: { Authorization: `Bearer ${token}` },
                });
                const recData = await recRes.json();
                setRecurring(Array.isArray(recData) ? recData : []);
            } catch (err: any) {
                setError(err.message);
            } finally {
                setLoading(false);
            }
        }

        fetchData();
    }, [token, authChecked]);

    // Category filtering
    useEffect(() => {
        if (!categories || categories.length === 0) {
            // If no categories exist yet, and user is typing, show type input
            if (categoryInput) {
                setFilteredCategories([]);
                setCategoryId(null);
                setShowCategoryTypeInput(true);
                setShowDropdown(false);
            }
            return;
        }
        if (!categoryInput) {
            setFilteredCategories(categories);
            setCategoryId(null);
            setShowCategoryTypeInput(false);
            setShowDropdown(false);
            return;
        }
        const filtered = categories.filter((cat) =>
            cat.name.toLowerCase().startsWith(categoryInput.toLowerCase())
        );
        setFilteredCategories(filtered);

        // Check for exact match
        const exactMatch = categories.find(
            (cat) => cat.name.toLowerCase() === categoryInput.toLowerCase()
        );

        if (exactMatch) {
            setCategoryId(exactMatch.id);
            setShowCategoryTypeInput(false);
            // Don't auto-show dropdown for exact matches (user selected from dropdown)
            setShowDropdown(false);
        } else {
            setCategoryId(null);
            setShowCategoryTypeInput(true);
            // Only show dropdown if there are filtered results and no exact match
            if (filtered.length > 0) {
                setShowDropdown(true);
            } else {
                setShowDropdown(false);
            }
        }
    }, [categoryInput, categories]);

    // Handle clicks outside dropdown
    useEffect(() => {
        function handleClickOutside(event: MouseEvent) {
            if (
                dropdownRef.current &&
                !dropdownRef.current.contains(event.target as Node) &&
                categoryInputRef.current &&
                !categoryInputRef.current.contains(event.target as Node)
            ) {
                setShowDropdown(false);
            }
        }
        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, []);

    const handleCategorySelect = (category: Category) => {
        setCategoryInput(category.name);
        setCategoryId(category.id);
        setShowDropdown(false);
        setShowCategoryTypeInput(false);
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!token) return;

        if (!categoryInput.trim()) {
            alert('Category name is required');
            return;
        }

        if (!categoryId && !categoryType) {
            alert('Please select category type for new category');
            return;
        }

        setSubmitting(true);
        try {
            let usedCategoryId = categoryId;

            // Create category if it doesn't exist
            if (!usedCategoryId) {
                const catFormData = new FormData();
                catFormData.append('name', categoryInput.trim());
                catFormData.append('type', categoryType);

                const catRes = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/category/add`, {
                    method: 'POST',
                    headers: { Authorization: `Bearer ${token}` },
                    body: catFormData,
                });
                const catData = await catRes.json();

                if (!catData.success) {
                    throw new Error(catData.error || 'Failed to create category');
                }

                usedCategoryId = catData.category.id;
                setCategories(prev => [...prev, catData.category]);
            }

            const formData = new FormData();
            formData.append('category_id', String(usedCategoryId));
            formData.append('amount', amount);
            formData.append('description', description);
            formData.append('start_date', startDate);
            formData.append('recurrence', recurrence);

            if (editingId) {
                formData.append('id', String(editingId));
            }

            const endpoint = editingId ? '/recurring/edit' : '/recurring/add';
            const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}${endpoint}`, {
                method: 'POST',
                headers: { Authorization: `Bearer ${token}` },
                body: formData,
            });

            if (!response.ok) throw new Error('Failed to save recurring transaction');

            // Refresh list
            const recRes = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/recurring/list`, {
                headers: { Authorization: `Bearer ${token}` },
            });
            const recData = await recRes.json();
            setRecurring(Array.isArray(recData) ? recData : []);

            // Reset form and close modal
            handleCloseModal();
        } catch (err: any) {
            alert('Failed to save: ' + err.message);
        } finally {
            setSubmitting(false);
        }
    };

    const handleOpenModal = () => {
        setShowModal(true);
    };

    const handleCloseModal = () => {
        setShowModal(false);
        setEditingId(null);
        setCategoryInput('');
        setCategoryId(null);
        setShowCategoryTypeInput(false);
        setCategoryType('expense');
        setAmount('');
        setDescription('');
        setStartDate(new Date().toISOString().slice(0, 10));
        setRecurrence('monthly');
    };

    const handleEdit = (rec: RecurringTransaction) => {
        const category = categories.find(c => c.id === rec.category_id);
        setEditingId(rec.id);
        setCategoryInput(category?.name || '');
        setCategoryId(rec.category_id);
        setAmount(String(rec.amount));
        setDescription(rec.description);
        setStartDate(rec.start_date.split('T')[0]);
        setRecurrence(rec.recurrence as any);
        setShowModal(true);
    };

    const handleDelete = async (id: number) => {
        if (!token) return;

        setDeleting(true);
        try {
            const formData = new FormData();
            formData.append('id', String(id));

            const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/recurring/delete`, {
                method: 'POST',
                headers: { Authorization: `Bearer ${token}` },
                body: formData,
            });

            if (!response.ok) throw new Error('Failed to delete');

            setRecurring(prev => prev.filter(r => r.id !== id));
            setDeleteConfirm(null);
        } catch (err: any) {
            alert('Failed to delete: ' + err.message);
        } finally {
            setDeleting(false);
        }
    };

    if (!authChecked || loading) {
        return (
            <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-indigo-50 via-white to-purple-50">
                <div className="text-center">
                    <div className="inline-block animate-spin rounded-full h-12 w-12 border-4 border-indigo-600 border-t-transparent mb-4"></div>
                    <p className="text-gray-600 font-medium">Loading...</p>
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
                        <div className="flex items-center space-x-2 sm:space-x-3">
                            <button
                                onClick={() => router.push('/dashboard')}
                                className="p-2 hover:bg-gray-100 rounded-lg transition-colors duration-200 flex-shrink-0"
                                aria-label="Back to dashboard"
                            >
                                <FiArrowLeft className="w-5 h-5 sm:w-6 sm:h-6 text-gray-700" />
                            </button>
                            <div className="flex items-center gap-1.5 sm:gap-2">
                                <FiRepeat className="w-6 h-6 sm:w-7 sm:h-7 text-indigo-600 flex-shrink-0" />
                                <h1 className="text-lg sm:text-2xl font-bold bg-gradient-to-r from-indigo-600 to-purple-600 bg-clip-text text-transparent">
                                    <span className="hidden sm:inline">Recurring Transactions</span>
                                    <span className="sm:hidden">Recurring</span>
                                </h1>
                            </div>
                        </div>
                    </div>
                </div>
            </header>

            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
                {/* Header Info */}
                <div className="mb-6">
                    <h2 className="text-xl sm:text-2xl font-bold text-gray-900 mb-2">Your Recurring Transactions</h2>
                    <p className="text-sm sm:text-base text-gray-600">Manage automatic recurring transactions</p>
                </div>

                {recurring.length === 0 ? (
                    <div className="bg-white/80 backdrop-blur-sm shadow-2xl rounded-2xl border border-white/20 p-8 sm:p-16 text-center">
                        <div className="w-20 h-20 bg-gradient-to-br from-indigo-100 to-purple-100 rounded-full flex items-center justify-center mx-auto mb-6">
                            <FiRepeat className="w-10 h-10 text-indigo-600" />
                        </div>
                        <h3 className="text-xl font-bold text-gray-900 mb-2">No recurring transactions yet</h3>
                        <p className="text-gray-500 mb-6">Set up automatic transactions for bills, subscriptions, or regular income</p>
                        <button
                            onClick={handleOpenModal}
                            className="inline-flex items-center gap-2 px-6 py-3 bg-gradient-to-r from-indigo-600 to-purple-600 text-white rounded-xl font-semibold hover:from-indigo-700 hover:to-purple-700 transition-all duration-200 shadow-lg hover:shadow-xl"
                        >
                            <FiPlus className="w-5 h-5" />
                            Add Your First Recurring Transaction
                        </button>
                    </div>
                ) : (
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                        {recurring.map((rec) => {
                            const category = categories.find(c => c.id === rec.category_id);
                            const isIncome = category?.type === 'income';

                            // Frequency colors matching dashboard style (soft backgrounds)
                            const frequencyConfig = {
                                daily: { bg: 'bg-slate-50', border: 'border-slate-200', text: 'text-slate-700', gradient: 'from-slate-500 to-gray-600', icon: '📅' },
                                weekly: { bg: 'bg-cyan-50', border: 'border-cyan-200', text: 'text-cyan-700', gradient: 'from-cyan-500 to-blue-600', icon: '📆' },
                                monthly: { bg: 'bg-indigo-50', border: 'border-indigo-200', text: 'text-indigo-700', gradient: 'from-indigo-500 to-purple-600', icon: '🗓️' },
                                yearly: { bg: 'bg-amber-50', border: 'border-amber-200', text: 'text-amber-700', gradient: 'from-amber-500 to-orange-600', icon: '📊' }
                            };
                            const config = frequencyConfig[rec.recurrence as keyof typeof frequencyConfig] || frequencyConfig.monthly;

                            return (
                                <div
                                    key={rec.id}
                                    className={`group bg-gradient-to-br ${config.bg} p-5 rounded-xl border-2 ${config.border} shadow-md hover:shadow-xl transition-all duration-200 transform hover:-translate-y-1`}
                                >
                                    <div className="flex items-start justify-between mb-4">
                                        <div className="flex-1">
                                            <h3 className="text-lg font-semibold text-gray-900 mb-2 line-clamp-2">
                                                {rec.description || 'No description'}
                                            </h3>
                                            <div className="flex items-center gap-2 mb-2">
                                                <span className={`text-xs px-2 py-1 bg-white rounded-full ${config.text} font-semibold capitalize`}>
                                                    {config.icon} {rec.recurrence}
                                                </span>
                                                <span className={`text-xs px-2 py-1 bg-white rounded-full font-semibold ${
                                                    isIncome ? 'text-emerald-700' : 'text-gray-700'
                                                }`}>
                                                    {category?.name || 'Unknown'}
                                                </span>
                                            </div>
                                        </div>
                                        <div className={`p-2 rounded-lg bg-gradient-to-br ${config.gradient} shadow-lg flex-shrink-0 group-hover:scale-110 transition-transform duration-200`}>
                                            <FiRepeat className="w-5 h-5 text-white" />
                                        </div>
                                    </div>

                                    <div className="mb-4">
                                        <p className={`text-3xl font-bold ${isIncome ? 'text-emerald-600' : 'text-gray-900'}`}>
                                            {isIncome && '+'}{!isIncome && ''}${rec.amount.toFixed(2)}
                                        </p>
                                        <p className="text-xs text-gray-600 mt-1">
                                            {isIncome ? '💰 Income' : '💳 Expense'}
                                        </p>
                                    </div>

                                    <div className="flex items-center justify-between pt-3 border-t border-gray-200">
                                        <div className="flex items-center gap-2 text-xs text-gray-600">
                                            <FiCalendar className="w-4 h-4" />
                                            <span>Starts {formatCalendarDate(rec.start_date)}</span>
                                        </div>
                                        <div className="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                                            <button
                                                onClick={() => handleEdit(rec)}
                                                className="p-1.5 bg-white hover:bg-gray-50 rounded-lg transition-colors shadow-sm"
                                                title="Edit"
                                            >
                                                <FiEdit2 className="w-4 h-4 text-gray-600" />
                                            </button>
                                            <button
                                                onClick={() => setDeleteConfirm(rec)}
                                                className="p-1.5 bg-white hover:bg-gray-50 rounded-lg transition-colors shadow-sm"
                                                title="Delete"
                                            >
                                                <FiTrash2 className="w-4 h-4 text-rose-600" />
                                            </button>
                                        </div>
                                    </div>
                                </div>
                            );
                        })}
                    </div>
                )}
            </div>

            {/* Floating Add Button */}
            <button
                onClick={handleOpenModal}
                aria-label="Add recurring transaction"
                className="fixed bottom-4 right-4 sm:bottom-6 sm:right-6 p-4 sm:p-5 rounded-full shadow-2xl bg-gradient-to-r from-indigo-600 to-purple-600 text-white hover:from-indigo-700 hover:to-purple-700 focus:outline-none focus:ring-4 focus:ring-indigo-500/50 transition-all duration-200 transform hover:scale-110 active:scale-95 z-50"
            >
                <FiPlus size={28} />
            </button>

            {/* Add/Edit Modal */}
            {showModal && (
                <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
                    <div className="bg-white rounded-2xl shadow-2xl max-w-md w-full max-h-[90vh] overflow-y-auto">
                        {/* Modal Header */}
                        <div className="sticky top-0 bg-white border-b border-gray-200 px-6 py-4 flex items-center justify-between rounded-t-2xl">
                            <h2 className="text-xl font-bold text-gray-900">
                                {editingId ? 'Edit Recurring Transaction' : 'Add Recurring Transaction'}
                            </h2>
                            <button
                                onClick={handleCloseModal}
                                className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
                                aria-label="Close"
                            >
                                <FiX className="w-5 h-5 text-gray-500" />
                            </button>
                        </div>

                        {/* Modal Body */}
                        <form onSubmit={handleSubmit} className="p-6 space-y-5">
                            {/* Category */}
                            <div>
                                <label className="block text-sm font-semibold text-gray-900 mb-2">
                                    🏷️ Category <span className="text-rose-500">*</span>
                                </label>
                                <div className="relative">
                                    <input
                                        ref={categoryInputRef}
                                        type="text"
                                        value={categoryInput}
                                        onChange={(e) => setCategoryInput(e.target.value)}
                                        onFocus={() => {
                                            if (categoryInput && filteredCategories.length > 0) {
                                                setShowDropdown(true);
                                            }
                                        }}
                                        className="w-full px-4 py-3 border-2 border-gray-300 rounded-xl focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 transition-all duration-200 bg-white text-gray-900 font-medium"
                                        placeholder="Type or select a category"
                                        autoComplete="off"
                                    />

                                    {/* Custom Dropdown */}
                                    {showDropdown && filteredCategories.length > 0 && (
                                        <div
                                            ref={dropdownRef}
                                            className="absolute z-20 w-full mt-2 bg-white border-2 border-gray-200 rounded-xl shadow-xl max-h-48 overflow-y-auto"
                                        >
                                            {filteredCategories.map((cat) => (
                                                <button
                                                    key={cat.id}
                                                    type="button"
                                                    onClick={() => handleCategorySelect(cat)}
                                                    className="w-full px-4 py-3 flex items-center justify-between hover:bg-gradient-to-r hover:from-indigo-50 hover:to-purple-50 transition-all duration-150 border-b border-gray-100 last:border-b-0 group text-left"
                                                >
                                                    <div className="flex items-center gap-3">
                                                        <div className={`w-2 h-2 rounded-full ${
                                                            cat.type === 'income' ? 'bg-emerald-400' : 'bg-rose-400'
                                                        }`} />
                                                        <span className="text-gray-900 font-medium group-hover:text-indigo-700 transition-colors">
                                                            {cat.name}
                                                        </span>
                                                    </div>
                                                    <span className={`text-xs font-semibold px-2.5 py-1 rounded-full ${
                                                        cat.type === 'income'
                                                            ? 'bg-emerald-50 text-emerald-700'
                                                            : 'bg-rose-50 text-rose-700'
                                                    }`}>
                                                        {cat.type === 'income' ? '💰 Income' : '💳 Expense'}
                                                    </span>
                                                </button>
                                            ))}
                                        </div>
                                    )}
                                </div>
                            </div>

                            {/* Category Type (for new categories) */}
                            {showCategoryTypeInput && (
                                <div className="bg-gradient-to-r from-indigo-50 to-purple-50 border-2 border-indigo-100 rounded-xl p-4">
                                    <label className="block text-sm font-semibold text-gray-700 mb-3">
                                        💡 New Category Type <span className="text-rose-500">*</span>
                                    </label>
                                    <p className="text-xs text-gray-600 mb-3">This category doesn't exist yet. Choose whether it's an income or expense.</p>
                                    <div className="grid grid-cols-2 gap-3">
                                        <button
                                            type="button"
                                            onClick={() => setCategoryType('expense')}
                                            className={`px-4 py-3.5 rounded-xl font-semibold transition-all duration-200 flex items-center justify-center gap-2 ${
                                                categoryType === 'expense'
                                                    ? 'bg-gradient-to-r from-rose-400 to-pink-400 text-white shadow-lg scale-105'
                                                    : 'bg-white text-gray-700 hover:bg-gray-50 border-2 border-gray-200'
                                            }`}
                                        >
                                            <span>💳</span> Expense
                                        </button>
                                        <button
                                            type="button"
                                            onClick={() => setCategoryType('income')}
                                            className={`px-4 py-3.5 rounded-xl font-semibold transition-all duration-200 flex items-center justify-center gap-2 ${
                                                categoryType === 'income'
                                                    ? 'bg-gradient-to-r from-emerald-400 to-teal-400 text-white shadow-lg scale-105'
                                                    : 'bg-white text-gray-700 hover:bg-gray-50 border-2 border-gray-200'
                                            }`}
                                        >
                                            <span>💰</span> Income
                                        </button>
                                    </div>
                                </div>
                            )}

                            {/* Amount */}
                            <div>
                                <label className="block text-sm font-semibold text-gray-900 mb-2">
                                    💵 Amount <span className="text-rose-500">*</span>
                                </label>
                                <div className="relative">
                                    <span className="absolute left-4 top-1/2 -translate-y-1/2 text-gray-500 font-semibold">$</span>
                                    <input
                                        type="number"
                                        min="0.01"
                                        step="0.01"
                                        value={amount}
                                        onChange={(e) => setAmount(e.target.value)}
                                        required
                                        className="w-full pl-8 pr-4 py-3 border-2 border-gray-300 rounded-xl focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 transition-all duration-200 text-gray-900 font-medium"
                                        placeholder="0.00"
                                    />
                                </div>
                            </div>

                            {/* Description */}
                            <div>
                                <label className="block text-sm font-semibold text-gray-900 mb-2">
                                    📝 Description <span className="text-gray-400 font-normal">(Optional)</span>
                                </label>
                                <input
                                    type="text"
                                    value={description}
                                    onChange={(e) => setDescription(e.target.value)}
                                    className="w-full px-4 py-3 border-2 border-gray-300 rounded-xl focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 transition-all duration-200 text-gray-900"
                                    placeholder="Optional description"
                                />
                            </div>

                            {/* Start Date */}
                            <div>
                                <label className="block text-sm font-semibold text-gray-900 mb-2">
                                    📅 Start Date <span className="text-rose-500">*</span>
                                </label>
                                <input
                                    type="date"
                                    value={startDate}
                                    onChange={(e) => setStartDate(e.target.value)}
                                    required
                                    className="w-full px-4 py-3 border-2 border-gray-300 rounded-xl focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 transition-all duration-200 text-gray-900 font-medium"
                                />
                            </div>

                            {/* Recurrence */}
                            <div>
                                <label className="block text-sm font-semibold text-gray-900 mb-2">
                                    🔄 Frequency <span className="text-rose-500">*</span>
                                </label>
                                <div className="grid grid-cols-2 gap-3">
                                    {(['daily', 'weekly', 'monthly', 'yearly'] as const).map((freq) => (
                                        <button
                                            key={freq}
                                            type="button"
                                            onClick={() => setRecurrence(freq)}
                                            className={`px-4 py-3 rounded-xl font-semibold capitalize transition-all duration-200 ${
                                                recurrence === freq
                                                    ? 'bg-gradient-to-r from-indigo-600 to-purple-600 text-white shadow-lg'
                                                    : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                                            }`}
                                        >
                                            {freq}
                                        </button>
                                    ))}
                                </div>
                            </div>

                            {/* Buttons */}
                            <div className="flex gap-3 pt-4">
                                <button
                                    type="button"
                                    onClick={handleCloseModal}
                                    className="flex-1 px-4 py-3 rounded-xl border-2 border-gray-300 text-gray-700 font-semibold hover:bg-gray-50 transition-all duration-200"
                                >
                                    Cancel
                                </button>
                                <button
                                    type="submit"
                                    disabled={submitting}
                                    className="flex-1 px-4 py-3 rounded-xl font-semibold text-white bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-700 hover:to-purple-700 transition-all duration-200 shadow-lg hover:shadow-xl disabled:opacity-50"
                                >
                                    {submitting ? 'Saving...' : (editingId ? 'Update' : 'Add')}
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            )}

            {/* Delete Confirmation Modal */}
            {deleteConfirm && (
                <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
                    <div className="bg-white rounded-2xl shadow-2xl max-w-md w-full p-6">
                        <div className="flex items-center gap-4 mb-4">
                            <div className="p-3 rounded-full bg-red-100">
                                <FiTrash2 className="w-6 h-6 text-red-600" />
                            </div>
                            <div>
                                <h3 className="text-xl font-bold text-gray-900">Delete Recurring Transaction</h3>
                                <p className="text-sm text-gray-600">This action cannot be undone</p>
                            </div>
                        </div>
                        <div className="mb-6 p-4 bg-gray-50 rounded-xl border border-gray-200">
                            <p className="text-sm font-semibold text-gray-900 mb-1">
                                {deleteConfirm.description || 'No description'}
                            </p>
                            <p className="text-sm text-gray-600">
                                ${deleteConfirm.amount.toFixed(2)} - {deleteConfirm.recurrence}
                            </p>
                        </div>
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
        </div>
    );
}
