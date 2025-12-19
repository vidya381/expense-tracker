'use client';

import { useState, useEffect, useRef } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '../../context/AuthContext';
import { FiRepeat, FiEdit2, FiTrash2, FiPlus, FiArrowLeft, FiCalendar, FiX } from 'react-icons/fi';
import { format, parseISO } from 'date-fns';

interface Category {
    id: number;
    name: string;
    type: 'income' | 'expense';
}

interface RecurringTransaction {
    ID: number;
    UserID: number;
    CategoryID: number;
    Amount: number;
    Description: string;
    StartDate: string;
    Recurrence: string;
    LastOccurrence: string | null;
    CreatedAt: string;
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

        if (filtered.length > 0) {
            setShowDropdown(true);
        } else {
            setShowDropdown(false);
        }

        const exactMatch = categories.find(
            (cat) => cat.name.toLowerCase() === categoryInput.toLowerCase()
        );

        if (exactMatch) {
            setCategoryId(exactMatch.id);
            setShowCategoryTypeInput(false);
        } else {
            setCategoryId(null);
            setShowCategoryTypeInput(true);
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
        const category = categories.find(c => c.id === rec.CategoryID);
        setEditingId(rec.ID);
        setCategoryInput(category?.name || '');
        setCategoryId(rec.CategoryID);
        setAmount(String(rec.Amount));
        setDescription(rec.Description);
        setStartDate(rec.StartDate);
        setRecurrence(rec.Recurrence as any);
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

            setRecurring(prev => prev.filter(r => r.ID !== id));
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
                        <div className="flex items-center space-x-3">
                            <button
                                onClick={() => router.push('/dashboard')}
                                className="p-2 hover:bg-gray-100 rounded-lg transition-colors duration-200"
                                aria-label="Back to dashboard"
                            >
                                <FiArrowLeft className="w-6 h-6 text-gray-700" />
                            </button>
                            <div className="flex items-center gap-2">
                                <FiRepeat className="w-7 h-7 text-indigo-600" />
                                <h1 className="text-2xl font-bold bg-gradient-to-r from-indigo-600 to-purple-600 bg-clip-text text-transparent">
                                    Recurring Transactions
                                </h1>
                            </div>
                        </div>
                    </div>
                </div>
            </header>

            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
                {/* Header Info */}
                <div className="mb-6">
                    <h2 className="text-2xl font-bold text-gray-900 mb-2">Your Recurring Transactions</h2>
                    <p className="text-gray-600">Manage automatic recurring transactions</p>
                </div>

                {recurring.length === 0 ? (
                    <div className="bg-white/80 backdrop-blur-sm shadow-2xl rounded-2xl border border-white/20 p-16 text-center">
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
                            const category = categories.find(c => c.id === rec.CategoryID);
                            const isIncome = category?.type === 'income';

                            // Frequency icon and color (balanced, vibrant but not harsh)
                            const frequencyConfig = {
                                daily: { bgColor: 'bg-blue-500', textColor: 'text-white', borderColor: 'border-blue-500', hoverBg: 'hover:bg-blue-600', icon: '📅' },
                                weekly: { bgColor: 'bg-teal-500', textColor: 'text-white', borderColor: 'border-teal-500', hoverBg: 'hover:bg-teal-600', icon: '📆' },
                                monthly: { bgColor: 'bg-indigo-500', textColor: 'text-white', borderColor: 'border-indigo-500', hoverBg: 'hover:bg-indigo-600', icon: '🗓️' },
                                yearly: { bgColor: 'bg-orange-500', textColor: 'text-white', borderColor: 'border-orange-500', hoverBg: 'hover:bg-orange-600', icon: '📊' }
                            };
                            const config = frequencyConfig[rec.Recurrence as keyof typeof frequencyConfig] || frequencyConfig.monthly;

                            return (
                                <div key={rec.ID} className={`group bg-white rounded-2xl border-2 ${config.borderColor} hover:border-gray-400 hover:shadow-xl transition-all duration-300 overflow-hidden`}>
                                    {/* Header with frequency badge */}
                                    <div className={`${config.bgColor} p-4`}>
                                        <div className="flex items-center justify-between">
                                            <div className="flex items-center gap-2">
                                                <span className="text-2xl">{config.icon}</span>
                                                <span className={`${config.textColor} font-bold text-sm uppercase tracking-wide`}>
                                                    {rec.Recurrence}
                                                </span>
                                            </div>
                                            <div className="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity duration-200">
                                                <button
                                                    onClick={() => handleEdit(rec)}
                                                    className="p-2 rounded-lg bg-white/30 hover:bg-white/50 text-white transition-all duration-150"
                                                    title="Edit"
                                                >
                                                    <FiEdit2 className="w-4 h-4" />
                                                </button>
                                                <button
                                                    onClick={() => setDeleteConfirm(rec)}
                                                    className="p-2 rounded-lg bg-white/30 hover:bg-white/50 text-white transition-all duration-150"
                                                    title="Delete"
                                                >
                                                    <FiTrash2 className="w-4 h-4" />
                                                </button>
                                            </div>
                                        </div>
                                    </div>

                                    {/* Content */}
                                    <div className="p-5">
                                        {/* Amount */}
                                        <div className="mb-4">
                                            <div className={`text-3xl font-bold ${isIncome ? 'text-green-600' : 'text-gray-900'}`}>
                                                {isIncome && '+'}{!isIncome && ''}${rec.Amount.toFixed(2)}
                                            </div>
                                        </div>

                                        {/* Description */}
                                        <h3 className="text-lg font-semibold text-gray-900 mb-3 line-clamp-2">
                                            {rec.Description || 'No description'}
                                        </h3>

                                        {/* Category */}
                                        <div className="mb-3">
                                            <span className={`inline-flex items-center px-3 py-1.5 rounded-lg text-sm font-semibold ${
                                                isIncome
                                                    ? 'bg-green-100 text-green-700'
                                                    : 'bg-gray-100 text-gray-700'
                                            }`}>
                                                {category?.name || 'Unknown'}
                                            </span>
                                        </div>

                                        {/* Start Date */}
                                        <div className="flex items-center gap-2 text-sm text-gray-600">
                                            <FiCalendar className="w-4 h-4" />
                                            <span>Starts {format(parseISO(rec.StartDate), 'MMM dd, yyyy')}</span>
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
                className="fixed bottom-6 right-6 p-5 rounded-full shadow-2xl bg-gradient-to-r from-indigo-600 to-purple-600 text-white hover:from-indigo-700 hover:to-purple-700 focus:outline-none focus:ring-4 focus:ring-indigo-500/50 transition-all duration-200 transform hover:scale-110 active:scale-95 z-50"
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
                                    Category <span className="text-red-500">*</span>
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
                                                            cat.type === 'income' ? 'bg-green-500' : 'bg-red-500'
                                                        }`} />
                                                        <span className="text-gray-900 font-medium group-hover:text-indigo-700 transition-colors">
                                                            {cat.name}
                                                        </span>
                                                    </div>
                                                    <span className={`text-xs font-semibold px-2.5 py-1 rounded-full ${
                                                        cat.type === 'income'
                                                            ? 'bg-green-100 text-green-700'
                                                            : 'bg-red-100 text-red-700'
                                                    }`}>
                                                        {cat.type === 'income' ? '+ Income' : '- Expense'}
                                                    </span>
                                                </button>
                                            ))}
                                        </div>
                                    )}
                                </div>
                            </div>

                            {/* Category Type (for new categories) */}
                            {showCategoryTypeInput && (
                                <div>
                                    <label className="block text-sm font-semibold text-gray-900 mb-2">
                                        Category Type <span className="text-red-500">*</span>
                                    </label>
                                    <div className="grid grid-cols-2 gap-3">
                                        <button
                                            type="button"
                                            onClick={() => setCategoryType('expense')}
                                            className={`px-4 py-3 rounded-xl font-semibold transition-all duration-200 ${
                                                categoryType === 'expense'
                                                    ? 'bg-gradient-to-r from-red-600 to-red-700 text-white shadow-lg'
                                                    : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                                            }`}
                                        >
                                            Expense
                                        </button>
                                        <button
                                            type="button"
                                            onClick={() => setCategoryType('income')}
                                            className={`px-4 py-3 rounded-xl font-semibold transition-all duration-200 ${
                                                categoryType === 'income'
                                                    ? 'bg-gradient-to-r from-green-600 to-green-700 text-white shadow-lg'
                                                    : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                                            }`}
                                        >
                                            Income
                                        </button>
                                    </div>
                                </div>
                            )}

                            {/* Amount */}
                            <div>
                                <label className="block text-sm font-semibold text-gray-900 mb-2">
                                    Amount <span className="text-red-500">*</span>
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
                                    Description
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
                                    Start Date <span className="text-red-500">*</span>
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
                                    Frequency <span className="text-red-500">*</span>
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
                                {deleteConfirm.Description || 'No description'}
                            </p>
                            <p className="text-sm text-gray-600">
                                ${deleteConfirm.Amount.toFixed(2)} - {deleteConfirm.Recurrence}
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
                                onClick={() => handleDelete(deleteConfirm.ID)}
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
