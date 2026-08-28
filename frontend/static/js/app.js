/**
 * FLOW — PERSONAL ACTIVITY TRACKER
 * Vanilla ES6+ Application Controller
 */

document.addEventListener('DOMContentLoaded', () => {
    // --- STATE ---
    let currentUser = null;
    let categories = [];
    let activities = [];
    let activeTimer = null;
    let timerInterval = null;

    // --- DOM ELEMENTS ---
    // Screens
    const authScreen = document.getElementById('auth-screen');
    const mainScreen = document.getElementById('main-screen');
    const userEmailDisplay = document.getElementById('user-email-display');

    // Auth Elements
    const authForm = document.getElementById('auth-form');
    const authEmail = document.getElementById('auth-email');
    const authPassword = document.getElementById('auth-password');
    const authSubmitBtn = document.getElementById('auth-submit-btn');
    const tabLoginBtn = document.getElementById('tab-login-btn');
    const tabRegisterBtn = document.getElementById('tab-register-btn');
    const logoutBtn = document.getElementById('logout-btn');
    let isRegisterMode = false;

    // Timer Elements
    const timerIdleView = document.getElementById('timer-idle-view');
    const timerActiveView = document.getElementById('timer-active-view');
    const timerStatusBadge = document.getElementById('timer-status-badge');
    const timerStartForm = document.getElementById('timer-start-form');
    const timerDescInput = document.getElementById('timer-desc');
    const timerCategorySelect = document.getElementById('timer-category-select');
    const activeCategoryPill = document.getElementById('active-category-pill');
    const activeDescriptionDisplay = document.getElementById('active-description-display');
    const activeTimerClock = document.getElementById('active-timer-clock');
    const stopTimerBtn = document.getElementById('stop-timer-btn');
    const discardTimerBtn = document.getElementById('discard-timer-btn');

    // Stats & Analytics Elements
    const statToday = document.getElementById('stat-today');
    const statWeek = document.getElementById('stat-week');
    const categoryDistributionContainer = document.getElementById('category-distribution-container');
    const weeklyChartContainer = document.getElementById('weekly-chart-container');

    // Activities Elements
    const historyCountBadge = document.getElementById('history-count-badge');
    const activityTableBody = document.getElementById('activity-table-body');
    const filterSearch = document.getElementById('filter-search');
    const filterCategory = document.getElementById('filter-category');
    const filterFromDate = document.getElementById('filter-from-date');
    const filterToDate = document.getElementById('filter-to-date');
    const filterFavoriteOnly = document.getElementById('filter-favorite-only');
    const clearFiltersBtn = document.getElementById('clear-filters-btn');
    const openManualLogBtn = document.getElementById('open-manual-log-btn');

    // Activity Modal Elements
    const activityModal = document.getElementById('activity-modal');
    const activityModalTitle = document.getElementById('activity-modal-title');
    const activityModalForm = document.getElementById('activity-modal-form');
    const modalActivityId = document.getElementById('modal-activity-id');
    const modalDescription = document.getElementById('modal-description');
    const modalCategory = document.getElementById('modal-category');
    const modalDate = document.getElementById('modal-date');
    const modalHours = document.getElementById('modal-hours');
    const modalMinutes = document.getElementById('modal-minutes');
    const modalNote = document.getElementById('modal-note');
    const modalFavorite = document.getElementById('modal-favorite');

    // Categories Modal Elements
    const categoriesModal = document.getElementById('categories-modal');
    const manageCategoriesBtn = document.getElementById('manage-categories-btn');
    const createCategoryForm = document.getElementById('create-category-form');
    const newCatName = document.getElementById('new-cat-name');
    const newCatColor = document.getElementById('new-cat-color');
    const categoryMgmtList = document.getElementById('category-mgmt-list');

    // --- API CLIENT HELPERS ---
    async function apiRequest(endpoint, options = {}) {
        const defaultHeaders = {
            'Content-Type': 'application/json',
        };

        const config = {
            ...options,
            headers: {
                ...defaultHeaders,
                ...options.headers,
            },
        };

        try {
            const response = await fetch(endpoint, config);
            if (response.status === 204) {
                return null;
            }
            const data = await response.json();
            if (!response.ok) {
                const error = new Error(data.error || 'An error occurred');
                error.status = response.status;
                error.code = data.code;
                throw error;
            }
            return data;
        } catch (err) {
            throw err;
        }
    }

    function showToast(message, type = 'success') {
        const container = document.getElementById('toast-container');
        const toast = document.createElement('div');
        toast.className = `toast toast-${type}`;
        toast.textContent = message;
        container.appendChild(toast);
        setTimeout(() => {
            toast.remove();
        }, 3500);
    }

    // --- AUTHENTICATION FLOWS ---
    tabLoginBtn.addEventListener('click', () => {
        isRegisterMode = false;
        tabLoginBtn.classList.add('active');
        tabRegisterBtn.classList.remove('active');
        authSubmitBtn.textContent = 'Log In';
    });

    tabRegisterBtn.addEventListener('click', () => {
        isRegisterMode = true;
        tabRegisterBtn.classList.add('active');
        tabLoginBtn.classList.remove('active');
        authSubmitBtn.textContent = 'Register';
    });

    authForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const email = authEmail.value.trim();
        const password = authPassword.value;

        if (!email || password.length < 8) {
            showToast('Password must be at least 8 characters long', 'error');
            return;
        }

        try {
            const endpoint = isRegisterMode ? '/api/auth/register' : '/api/auth/login';
            const data = await apiRequest(endpoint, {
                method: 'POST',
                body: JSON.stringify({ email, password }),
            });

            currentUser = data.user;
            showToast(isRegisterMode ? 'Registration successful!' : 'Welcome back!');
            showMainScreen();
        } catch (err) {
            showToast(err.message, 'error');
        }
    });

    logoutBtn.addEventListener('click', async () => {
        try {
            await apiRequest('/api/auth/logout', { method: 'POST' });
        } catch (e) {
            // ignore
        }
        currentUser = null;
        clearInterval(timerInterval);
        showAuthScreen();
        showToast('Logged out successfully');
    });

    async function checkAuth() {
        try {
            const data = await apiRequest('/api/auth/me');
            currentUser = data.user;
            showMainScreen();
        } catch (err) {
            showAuthScreen();
        }
    }

    function showAuthScreen() {
        authScreen.classList.remove('hidden');
        mainScreen.classList.add('hidden');
    }

    async function showMainScreen() {
        authScreen.classList.add('hidden');
        mainScreen.classList.remove('hidden');
        userEmailDisplay.textContent = currentUser.email;

        // Load all core data
        await loadCategories();
        await loadActiveTimer();
        await loadDashboard();
        await loadActivities();
    }

    // --- CATEGORIES ---
    async function loadCategories() {
        try {
            const data = await apiRequest('/api/categories');
            categories = data.categories || [];
            renderCategoryDropdowns();
            renderCategoryManagementList();
        } catch (err) {
            showToast('Failed to load categories', 'error');
        }
    }

    function renderCategoryDropdowns() {
        const populate = (selectElement, includeAllOption = false) => {
            selectElement.innerHTML = '';
            if (includeAllOption) {
                const opt = document.createElement('option');
                opt.value = '';
                opt.textContent = 'All Categories';
                selectElement.appendChild(opt);
            }
            categories.forEach((cat) => {
                const opt = document.createElement('option');
                opt.value = cat.id;
                opt.textContent = cat.name;
                selectElement.appendChild(opt);
            });
        };

        populate(timerCategorySelect, false);
        populate(filterCategory, true);
        populate(modalCategory, false);
    }

    function renderCategoryManagementList() {
        categoryMgmtList.innerHTML = '';
        if (categories.length === 0) {
            categoryMgmtList.innerHTML = '<p class="empty-text">No categories created yet.</p>';
            return;
        }

        categories.forEach((cat) => {
            const item = document.createElement('div');
            item.className = 'category-mgmt-item';

            const left = document.createElement('div');
            left.className = 'category-badge-group';

            const dot = document.createElement('div');
            dot.className = 'color-dot';
            dot.style.backgroundColor = cat.color;

            const name = document.createElement('span');
            name.textContent = cat.name;
            name.style.fontWeight = '600';

            left.appendChild(dot);
            left.appendChild(name);

            const deleteBtn = document.createElement('button');
            deleteBtn.type = 'button';
            deleteBtn.className = 'btn-icon';
            deleteBtn.innerHTML = '🗑️';
            deleteBtn.title = 'Delete category';
            deleteBtn.addEventListener('click', () => deleteCategory(cat.id));

            item.appendChild(left);
            item.appendChild(deleteBtn);
            categoryMgmtList.appendChild(item);
        });
    }

    createCategoryForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const name = newCatName.value.trim();
        const color = newCatColor.value;
        if (!name) return;

        try {
            await apiRequest('/api/categories', {
                method: 'POST',
                body: JSON.stringify({ name, color }),
            });
            newCatName.value = '';
            showToast('Category created');
            await loadCategories();
            await loadDashboard();
        } catch (err) {
            showToast(err.message, 'error');
        }
    });

    async function deleteCategory(id) {
        if (!confirm('Are you sure you want to delete this category?')) return;
        try {
            await apiRequest(`/api/categories/${id}`, { method: 'DELETE' });
            showToast('Category deleted');
            await loadCategories();
            await loadDashboard();
            await loadActivities();
        } catch (err) {
            // Handles Rule 4 (Category Protection)
            showToast(err.message, 'error');
        }
    }

    // --- PERSISTENT TIMER FLOW ---
    async function loadActiveTimer() {
        try {
            const data = await apiRequest('/api/timer');
            if (data.active_timer) {
                activeTimer = data.active_timer;
                startTimerTicker();
                showActiveTimerView();
            } else {
                activeTimer = null;
                clearInterval(timerInterval);
                showIdleTimerView();
            }
        } catch (err) {
            showIdleTimerView();
        }
    }

    function showIdleTimerView() {
        timerIdleView.classList.remove('hidden');
        timerActiveView.classList.add('hidden');
        timerStatusBadge.className = 'badge badge-idle';
        timerStatusBadge.textContent = 'Idle';
    }

    function showActiveTimerView() {
        timerIdleView.classList.add('hidden');
        timerActiveView.classList.remove('hidden');
        timerStatusBadge.className = 'badge badge-active';
        timerStatusBadge.textContent = 'Tracking';

        activeDescriptionDisplay.textContent = activeTimer.description;
        activeCategoryPill.textContent = activeTimer.category_name || 'Category';
        activeCategoryPill.style.backgroundColor = (activeTimer.category_color || '#4F46E5') + '22';
        activeCategoryPill.style.color = activeTimer.category_color || '#4F46E5';
    }

    function startTimerTicker() {
        clearInterval(timerInterval);
        const startedAtMs = new Date(activeTimer.started_at).getTime();

        const updateClock = () => {
            const nowMs = Date.now();
            const elapsedSec = Math.max(0, Math.floor((nowMs - startedAtMs) / 1000));

            const hrs = Math.floor(elapsedSec / 3600);
            const mins = Math.floor((elapsedSec % 3600) / 60);
            const secs = elapsedSec % 60;

            activeTimerClock.textContent = `${pad(hrs)}:${pad(mins)}:${pad(secs)}`;
        };

        updateClock();
        timerInterval = setInterval(updateClock, 1000);
    }

    function pad(num) {
        return num.toString().padStart(2, '0');
    }

    timerStartForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const description = timerDescInput.value.trim();
        const category_id = timerCategorySelect.value;

        if (!description || !category_id) {
            showToast('Please provide a description and category', 'error');
            return;
        }

        try {
            const data = await apiRequest('/api/timer/start', {
                method: 'POST',
                body: JSON.stringify({ description, category_id }),
            });

            activeTimer = data.active_timer;
            timerDescInput.value = '';
            showToast('Timer started');
            startTimerTicker();
            showActiveTimerView();
        } catch (err) {
            showToast(err.message, 'error');
        }
    });

    stopTimerBtn.addEventListener('click', async () => {
        try {
            await apiRequest('/api/timer/stop', {
                method: 'POST',
                body: JSON.stringify({}),
            });

            clearInterval(timerInterval);
            activeTimer = null;
            showIdleTimerView();
            showToast('Activity logged successfully!');

            // Behavior 3: Refresh dashboard and activities
            await loadDashboard();
            await loadActivities();
        } catch (err) {
            showToast(err.message, 'error');
        }
    });

    discardTimerBtn.addEventListener('click', async () => {
        if (!confirm('Discard this ongoing timer without saving?')) return;
        try {
            await apiRequest('/api/timer/discard', { method: 'DELETE' });
            clearInterval(timerInterval);
            activeTimer = null;
            showIdleTimerView();
            showToast('Timer discarded');
        } catch (err) {
            showToast(err.message, 'error');
        }
    });

    // --- DASHBOARD & ANALYTICS ---
    async function loadDashboard() {
        try {
            const tzOffset = new Date().getTimezoneOffset();
            const data = await apiRequest(`/api/dashboard?tz_offset=${tzOffset}`);

            statToday.textContent = data.today_formatted || '0m';
            statWeek.textContent = data.week_formatted || '0m';

            renderCategoryDistribution(data.category_distribution || []);
            renderWeeklyChart(data.daily_activity || []);
        } catch (err) {
            showToast('Failed to load dashboard metrics', 'error');
        }
    }

    function renderCategoryDistribution(distList) {
        categoryDistributionContainer.innerHTML = '';
        if (distList.length === 0) {
            categoryDistributionContainer.innerHTML = '<p class="empty-text">No activities logged this week yet.</p>';
            return;
        }

        distList.forEach((item) => {
            const row = document.createElement('div');
            row.className = 'distribution-item';

            const header = document.createElement('div');
            header.className = 'distribution-item-header';

            const name = document.createElement('span');
            name.textContent = item.name;

            const percentage = document.createElement('span');
            percentage.textContent = `${item.percentage}% (${formatSeconds(item.seconds)})`;
            percentage.style.color = 'var(--text-secondary)';

            header.appendChild(name);
            header.appendChild(percentage);

            const barBg = document.createElement('div');
            barBg.className = 'progress-bar-bg';

            const barFill = document.createElement('div');
            barFill.className = 'progress-bar-fill';
            barFill.style.width = `${Math.min(100, item.percentage)}%`;
            barFill.style.backgroundColor = item.color || '#4F46E5';

            barBg.appendChild(barFill);
            row.appendChild(header);
            row.appendChild(barBg);
            categoryDistributionContainer.appendChild(row);
        });
    }

    function renderWeeklyChart(dailyList) {
        weeklyChartContainer.innerHTML = '';
        if (dailyList.length === 0) return;

        const maxSeconds = Math.max(...dailyList.map((d) => d.seconds), 3600);

        dailyList.forEach((d) => {
            const col = document.createElement('div');
            col.className = 'chart-col';

            const barWrap = document.createElement('div');
            barWrap.className = 'chart-bar-wrap';

            const bar = document.createElement('div');
            bar.className = 'chart-bar';
            const heightPct = (d.seconds / maxSeconds) * 100;
            bar.style.height = `${Math.max(4, heightPct)}%`;
            if (d.seconds === 0) {
                bar.style.opacity = '0.2';
            }
            bar.title = `${d.day} (${d.date}): ${formatSeconds(d.seconds)}`;

            barWrap.appendChild(bar);

            const label = document.createElement('span');
            label.className = 'chart-day';
            label.textContent = d.day;

            col.appendChild(barWrap);
            col.appendChild(label);
            weeklyChartContainer.appendChild(col);
        });
    }

    // --- ACTIVITIES LIST & FILTERING ---
    async function loadActivities() {
        try {
            const params = new URLSearchParams();
            if (filterSearch.value.trim()) params.set('search', filterSearch.value.trim());
            if (filterCategory.value) params.set('category_id', filterCategory.value);
            if (filterFromDate.value) params.set('from_date', filterFromDate.value);
            if (filterToDate.value) params.set('to_date', filterToDate.value);
            if (filterFavoriteOnly.checked) params.set('favorite', 'true');

            const data = await apiRequest(`/api/activities?${params.toString()}`);
            activities = data.activities || [];
            historyCountBadge.textContent = `${activities.length} activities`;
            renderActivityTable();
        } catch (err) {
            showToast('Failed to load activity history', 'error');
        }
    }

    function renderActivityTable() {
        activityTableBody.innerHTML = '';
        if (activities.length === 0) {
            activityTableBody.innerHTML = '<tr><td colspan="7" class="text-center empty-state">No matching activities found.</td></tr>';
            return;
        }

        activities.forEach((act) => {
            const tr = document.createElement('tr');

            // 1. Star Favorite
            const tdFav = document.createElement('td');
            const star = document.createElement('span');
            star.className = `fav-star ${act.favorite ? 'active' : ''}`;
            star.textContent = '★';
            star.title = act.favorite ? 'Unmark favorite' : 'Mark as favorite';
            star.addEventListener('click', () => toggleFavorite(act.id, !act.favorite));
            tdFav.appendChild(star);

            // 2. Date
            const tdDate = document.createElement('td');
            tdDate.textContent = act.activity_date;

            // 3. Category
            const tdCat = document.createElement('td');
            const pill = document.createElement('span');
            pill.className = 'category-pill';
            pill.textContent = act.category_name;
            pill.style.backgroundColor = (act.category_color || '#4F46E5') + '22';
            pill.style.color = act.category_color || '#4F46E5';
            tdCat.appendChild(pill);

            // 4. Description
            const tdDesc = document.createElement('td');
            tdDesc.textContent = act.description;
            tdDesc.style.fontWeight = '500';

            // 5. Duration
            const tdDur = document.createElement('td');
            tdDur.textContent = formatSeconds(act.duration_seconds);
            tdDur.style.fontFamily = 'var(--font-mono)';

            // 6. Note
            const tdNote = document.createElement('td');
            tdNote.textContent = act.note || '—';
            tdNote.style.color = act.note ? 'var(--text-secondary)' : 'var(--text-muted)';

            // 7. Actions
            const tdActions = document.createElement('td');
            tdActions.className = 'text-right';

            const actionWrap = document.createElement('div');
            actionWrap.className = 'action-buttons';

            const editBtn = document.createElement('button');
            editBtn.type = 'button';
            editBtn.className = 'btn-icon';
            editBtn.innerHTML = '✏️';
            editBtn.title = 'Edit activity';
            editBtn.addEventListener('click', () => openEditModal(act));

            const delBtn = document.createElement('button');
            delBtn.type = 'button';
            delBtn.className = 'btn-icon';
            delBtn.innerHTML = '🗑️';
            delBtn.title = 'Delete activity';
            delBtn.addEventListener('click', () => deleteActivity(act.id));

            actionWrap.appendChild(editBtn);
            actionWrap.appendChild(delBtn);
            tdActions.appendChild(actionWrap);

            tr.appendChild(tdFav);
            tr.appendChild(tdDate);
            tr.appendChild(tdCat);
            tr.appendChild(tdDesc);
            tr.appendChild(tdDur);
            tr.appendChild(tdNote);
            tr.appendChild(tdActions);

            activityTableBody.appendChild(tr);
        });
    }

    async function toggleFavorite(id, favorite) {
        try {
            await apiRequest(`/api/activities/${id}/favorite`, {
                method: 'PATCH',
                body: JSON.stringify({ favorite }),
            });
            await loadActivities();
        } catch (err) {
            showToast('Failed to update favorite', 'error');
        }
    }

    async function deleteActivity(id) {
        if (!confirm('Are you sure you want to delete this activity?')) return;
        try {
            await apiRequest(`/api/activities/${id}`, { method: 'DELETE' });
            showToast('Activity deleted');
            await loadActivities();
            await loadDashboard();
        } catch (err) {
            showToast(err.message, 'error');
        }
    }

    // Behavior 4: Filter changes trigger reactive reload
    let searchDebounceTimer;
    filterSearch.addEventListener('input', () => {
        clearTimeout(searchDebounceTimer);
        searchDebounceTimer = setTimeout(loadActivities, 250);
    });

    filterCategory.addEventListener('change', loadActivities);
    filterFromDate.addEventListener('change', loadActivities);
    filterToDate.addEventListener('change', loadActivities);
    filterFavoriteOnly.addEventListener('change', loadActivities);

    clearFiltersBtn.addEventListener('click', () => {
        filterSearch.value = '';
        filterCategory.value = '';
        filterFromDate.value = '';
        filterToDate.value = '';
        filterFavoriteOnly.checked = false;
        loadActivities();
    });

    // --- MANUAL ACTIVITY LOG / EDIT MODAL ---
    openManualLogBtn.addEventListener('click', () => {
        openCreateModal();
    });

    function openCreateModal() {
        modalActivityId.value = '';
        activityModalTitle.textContent = 'Log Activity Manually';
        modalDescription.value = '';
        modalDate.value = new Date().toISOString().split('T')[0];
        modalHours.value = '1';
        modalMinutes.value = '0';
        modalNote.value = '';
        modalFavorite.checked = false;
        activityModal.classList.remove('hidden');
    }

    function openEditModal(act) {
        modalActivityId.value = act.id;
        activityModalTitle.textContent = 'Edit Activity';
        modalDescription.value = act.description;
        modalCategory.value = act.category_id;
        modalDate.value = act.activity_date;
        modalHours.value = Math.floor(act.duration_seconds / 3600);
        modalMinutes.value = Math.floor((act.duration_seconds % 3600) / 60);
        modalNote.value = act.note || '';
        modalFavorite.checked = act.favorite;
        activityModal.classList.remove('hidden');
    }

    activityModalForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const id = modalActivityId.value;
        const description = modalDescription.value.trim();
        const category_id = modalCategory.value;
        const activity_date = modalDate.value;
        const hours = parseInt(modalHours.value || '0', 10);
        const minutes = parseInt(modalMinutes.value || '0', 10);
        const duration_seconds = hours * 3600 + minutes * 60;
        const note = modalNote.value.trim() || null;
        const favorite = modalFavorite.checked;

        // Behavior 1: Client side validation
        if (!description || !category_id || !activity_date || duration_seconds <= 0) {
            showToast('Please enter a description, category, date, and duration > 0', 'error');
            return;
        }

        try {
            const body = { description, category_id, activity_date, duration_seconds, note, favorite };
            if (id) {
                await apiRequest(`/api/activities/${id}`, {
                    method: 'PUT',
                    body: JSON.stringify(body),
                });
                showToast('Activity updated');
            } else {
                await apiRequest('/api/activities', {
                    method: 'POST',
                    body: JSON.stringify(body),
                });
                showToast('Activity logged');
            }

            closeAllModals();
            await loadActivities();
            await loadDashboard();
        } catch (err) {
            showToast(err.message, 'error');
        }
    });

    // --- MODAL CONTROLS ---
    manageCategoriesBtn.addEventListener('click', () => {
        categoriesModal.classList.remove('hidden');
    });

    document.querySelectorAll('[data-close-modal]').forEach((btn) => {
        btn.addEventListener('click', closeAllModals);
    });

    function closeAllModals() {
        activityModal.classList.add('hidden');
        categoriesModal.classList.add('hidden');
    }

    function formatSeconds(seconds) {
        if (!seconds || seconds <= 0) return '0m';
        const hrs = Math.floor(seconds / 3600);
        const mins = Math.floor((seconds % 3600) / 60);
        if (hrs > 0) return `${hrs}h ${pad(mins)}m`;
        return `${mins}m`;
    }

    // --- STARTUP ---
    checkAuth();
});
