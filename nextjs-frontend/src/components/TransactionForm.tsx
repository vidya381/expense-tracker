'use client';

import React, { useState, useEffect, useRef, Dispatch, SetStateAction } from 'react';


interface Category {
    id: number;
    name: string;
    type: 'income' | 'expense';
}

interface TransactionFormProps {
    token: string;
    categories: Category[];
    setCategories: Dispatch<SetStateAction<Category[]>>;
    onSuccess: () => void;
    onCancel?: () => void;
    initialValues?: {
        id?: number;
        category_id?: number;
        category_name?: string;
        amount?: number;
        description?: string;
        date?: string;
    };
}

export default function TransactionForm({
    token,
    categories,
    setCategories,
    onSuccess,
    onCancel,
    initialValues = {},
}: TransactionFormProps) {

    const [filteredCategories, setFilteredCategories] = useState<Category[]>([]);

    const [categoryInput, setCategoryInput] = useState(
        initialValues.category_name || ''
    );
    const [categoryId, setCategoryId] = useState<number | null>(
        initialValues.category_id || null
    );

    const [showCategoryTypeInput, setShowCategoryTypeInput] = useState(false);
    const [categoryType, setCategoryType] = useState<'income' | 'expense'>('expense');
    const [amount, setAmount] = useState(
        initialValues.amount !== undefined ? initialValues.amount.toString() : ''
    );
    const [description, setDescription] = useState(initialValues.description || '');
    const [date, setDate] = useState(
        initialValues.date || new Date().toISOString().slice(0, 10)
    );

    const [loadingCategories, setLoadingCategories] = useState(false);
    const [loadingSubmit, setLoadingSubmit] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const categoryInputRef = useRef<HTMLInputElement>(null);

    useEffect(() => {
        async function fetchCategories() {
            setLoadingCategories(true);
            try {
                const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/category/list`, {
                    headers: { Authorization: `Bearer ${token}` },
                });
                const data = await res.json();
                if (data.success) {
                    setCategories(data.categories);
                    setFilteredCategories(data.categories);
                } else {
                    setError(data.error || 'Failed to load categories');
                }
            } catch (e: any) {
                setError(e.message);
            } finally {
                setLoadingCategories(false);
            }
        }
        fetchCategories();
    }, [token]);

    useEffect(() => {
        if (!categoryInput) {
            setFilteredCategories(categories);
            setCategoryId(null);
            setShowCategoryTypeInput(false);
            return;
        }
        const filtered = categories.filter((cat) =>
            cat.name.toLowerCase().startsWith(categoryInput.toLowerCase())
        );
        setFilteredCategories(filtered);

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

    async function handleSubmit(e: React.FormEvent) {
        e.preventDefault();
        setError(null);
        if (!categoryInput.trim()) {
            setError('Category name is required');
            categoryInputRef.current?.focus();
            return;
        }
        if (!categoryId && !categoryType) {
            setError('Please select category type for new category');
            return;
        }
        const numericAmount = parseFloat(amount);
        if (isNaN(numericAmount) || numericAmount <= 0) {
            setError("Enter a valid amount greater than zero.");
            return;
        }
        if (!date) {
            setError('Select a valid date');
            return;
        }

        setLoadingSubmit(true);
        try {
            let usedCategoryId = categoryId;
            if (!usedCategoryId) {
                const formData = new FormData();
                formData.append('name', categoryInput.trim());
                formData.append('type', categoryType);

                // const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/category/add`, {
                //     method: 'POST',
                //     headers: { Authorization: `Bearer ${token}` },
                //     body: formData,
                // });
                // const data = await res.json();
                // if (!data.success) throw new Error(data.error || 'Failed to create category');
                // usedCategoryId = data.category.id;
                // setCategories((prev) => [...prev, data.category]);
                const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/category/add`, {
                    method: 'POST',
                    headers: { Authorization: `Bearer ${token}` },
                    body: formData,
                });
                const data = await res.json();

                if (!data.success) {
                    throw new Error(data.error || 'Failed to create category');
                }

                const createdCategory = data.category;
                if (!createdCategory || !createdCategory.id) {
                    throw new Error('Invalid category data from server');
                }

                usedCategoryId = createdCategory.id;

                // Update local categories state
                // setCategories(prev => [...prev, createdCategory]);
                setCategories(prev => [...prev, data.category]);


            }

            const txForm = new FormData();
            txForm.append('category_id', String(usedCategoryId));
            txForm.append('amount', amount.toString());
            txForm.append('description', description);
            txForm.append('date', date);

            // If editing, append id
            if (initialValues.id) {
                txForm.append('id', String(initialValues.id));
            }

            const txUrl = initialValues.id
                ? `${process.env.NEXT_PUBLIC_API_URL}/transaction/update`
                : `${process.env.NEXT_PUBLIC_API_URL}/transaction/add`;

            const txRes = await fetch(txUrl, {
                method: 'POST',
                headers: { Authorization: `Bearer ${token}` },
                body: txForm,
            });
            const txData = await txRes.json();
            if (!txData.success) throw new Error(txData.error || 'Failed to save transaction');

            // Reset form if adding
            if (!initialValues.id) {
                setCategoryInput('');
                setCategoryId(null);
                setDescription('');
                setAmount('0');
                setDate(new Date().toISOString().slice(0, 10));
                setShowCategoryTypeInput(false);
            }
            onSuccess();
        } catch (e: any) {
            setError(e.message);
        } finally {
            setLoadingSubmit(false);
        }
    }

    return (
        <form className="max-w-md mx-auto space-y-4" onSubmit={handleSubmit}>
            <div>
                <label htmlFor="category" className="block font-medium mb-1">Category</label>
                <input
                    list="categorylist"
                    id="category"
                    value={categoryInput}
                    onChange={(e) => setCategoryInput(e.target.value)}
                    disabled={loadingSubmit || loadingCategories}
                    className="w-full border border-gray-300 rounded px-3 py-2"
                    autoComplete="off"
                    placeholder="Type or select a category"
                    ref={categoryInputRef}
                />
                <datalist id="categorylist">
                    {filteredCategories.map((cat) => (
                        <option key={cat.id} value={cat.name} />
                    ))}
                </datalist>
            </div>

            {showCategoryTypeInput && (
                <div>
                    <label className="block font-medium mb-1">Category Type</label>
                    <select
                        value={categoryType}
                        onChange={(e) => setCategoryType(e.target.value as 'income' | 'expense')}
                        disabled={loadingSubmit}
                        className="w-full border border-gray-300 rounded px-3 py-2"
                    >
                        <option value="expense">Expense</option>
                        <option value="income">Income</option>
                    </select>
                </div>
            )}

            <div>
                <label htmlFor="amount" className="block font-medium mb-1">Amount</label>
                <input
                    id="amount"
                    type="number"
                    min="0.01"
                    step="0.01"
                    value={amount}
                    onChange={e => setAmount(e.target.value)}
                    disabled={loadingSubmit}
                    className="w-full border border-gray-300 rounded px-3 py-2"
                    required
                />
            </div>

            <div>
                <label htmlFor="description" className="block font-medium mb-1">Description</label>
                <input
                    id="description"
                    type="text"
                    value={description}
                    onChange={(e) => setDescription(e.target.value)}
                    disabled={loadingSubmit}
                    className="w-full border border-gray-300 rounded px-3 py-2"
                    placeholder="Optional"
                />
            </div>

            <div>
                <label htmlFor="date" className="block font-medium mb-1">Date</label>
                <input
                    id="date"
                    type="date"
                    value={date}
                    onChange={(e) => setDate(e.target.value)}
                    disabled={loadingSubmit}
                    className="w-full border border-gray-300 rounded px-3 py-2"
                    required
                />
            </div>

            {error && <p className="text-red-600 font-semibold text-sm mt-1">{error}</p>}

            <button
                type="submit"
                disabled={loadingSubmit}
                className="w-full bg-indigo-600 hover:bg-indigo-700 text-white py-2 rounded font-semibold mt-4 disabled:opacity-50 disabled:cursor-not-allowed"
            >
                {loadingSubmit ? 'Saving...' : initialValues.id ? 'Update Transaction' : 'Add Transaction'}
            </button>
            {onCancel && (
                <button
                    type="button"
                    onClick={onCancel}
                    className="w-full mt-2 py-2 rounded border border-gray-300 hover:bg-gray-100"
                    disabled={loadingSubmit}
                >
                    Cancel
                </button>
            )}
        </form>
    );
}
