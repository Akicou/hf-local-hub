class App {
  constructor() {
    this.currentTab = 'models';
    this.token = localStorage.getItem('auth_token');
    this.repositories = [];
    this.filters = { type: 'all', privacy: 'all' };
    this.authMethods = { token: true, hf: false, ldap: false };
    this.init();
  }

  async init() {
    await this.fetchAuthConfig();
    this.bindNav();
    this.bindSidebar();
    this.bindModal();
    this.bindSearch();
    this.updateAuthUI();
    this.loadRepositories();
  }

  async fetchAuthConfig() {
    try {
      const res = await fetch('/api/auth/config');
      if (res.ok) {
        const cfg = await res.json();
        this.authMethods = cfg;
      }
    } catch {}
  }

  bindNav() {
    document.querySelectorAll('.nav-link[data-tab]').forEach(btn => {
      btn.addEventListener('click', () => this.switchTab(btn.dataset.tab));
    });
    document.getElementById('loginBtn').addEventListener('click', () =>
      this.token ? this.logout() : this.login()
    );
    document.getElementById('createRepoBtn').addEventListener('click', () => this.showModal());
  }

  bindSidebar() {
    document.querySelectorAll('.filter-pill').forEach(pill => {
      pill.addEventListener('click', () => {
        const { filter, value } = pill.dataset;
        this.filters[filter] = value;
        document.querySelectorAll(`.filter-pill[data-filter="${filter}"]`).forEach(p =>
          p.classList.toggle('active', p === pill)
        );
        this.renderRepositories();
      });
    });
  }

  bindModal() {
    const overlay = document.getElementById('repoModal');
    const close = () => {
      overlay.classList.remove('show');
      document.getElementById('createRepoForm').reset();
    };
    document.getElementById('modalClose').addEventListener('click', close);
    document.getElementById('modalCancel').addEventListener('click', close);
    overlay.addEventListener('click', e => { if (e.target === overlay) close(); });
    document.getElementById('modalSubmit').addEventListener('click', () => this.createRepo());
  }

  bindSearch() {
    document.getElementById('searchInput').addEventListener('input', e => {
      this.searchQuery = e.target.value;
      this.renderRepositories();
    });
  }

  switchTab(tab) {
    this.currentTab = tab;
    document.querySelectorAll('.nav-link[data-tab]').forEach(btn =>
      btn.classList.toggle('active', btn.dataset.tab === tab)
    );
    document.getElementById('pageTitle').textContent =
      tab.charAt(0).toUpperCase() + tab.slice(1);
    this.loadRepositories();
  }

  async loadRepositories() {
    const endpoint = this.currentTab === 'models' ? '/api/models/' : '/api/datasets/';
    try {
      const res = await fetch(endpoint);
      if (!res.ok) throw new Error();
      this.repositories = await res.json() || [];
    } catch {
      this.repositories = [];
      this.showToast('Failed to load repositories', 'error');
    }
    this.renderRepositories();
  }

  filtered() {
    return this.repositories.filter(r => {
      if (this.filters.type !== 'all' && r.type !== this.filters.type) return false;
      if (this.filters.privacy === 'public' && r.private) return false;
      if (this.filters.privacy === 'private' && !r.private) return false;
      const q = (this.searchQuery || '').toLowerCase();
      if (q && !r.name.toLowerCase().includes(q) && !r.namespace.toLowerCase().includes(q)) return false;
      return true;
    });
  }

  renderRepositories() {
    const list = this.filtered();
    const count = document.getElementById('repoCount');
    count.textContent = list.length.toLocaleString();

    const container = document.getElementById('repoList');
    if (list.length === 0) {
      container.innerHTML = `
        <div class="empty-state">
          <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/>
          </svg>
          <p>No repositories found</p>
        </div>`;
      return;
    }
    container.innerHTML = list.map(r => this.cardHTML(r)).join('');
  }

  cardHTML(r) {
    const icons = { model: '🤖', dataset: '📊', space: '🚀' };
    const badgeClass = { model: 'badge-model', dataset: 'badge-dataset', space: 'badge-space' };
    const iconClass = { model: 'icon-model', dataset: 'icon-dataset', space: 'icon-space' };
    const t = r.type || 'model';
    const date = r.created_at
      ? new Date(r.created_at).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
      : '';
    const repoId = r.repo_id || `${r.namespace}/${r.name}`;
    return `
      <a href="/r/${repoId}" class="repo-card" style="text-decoration:none;color:inherit">
        <div class="repo-card-top">
          <div class="repo-icon ${iconClass[t] || 'icon-model'}">${icons[t] || '🤖'}</div>
          <div class="repo-card-meta">
            <div class="repo-name"><span class="ns">${this.esc(r.namespace)}/</span>${this.esc(r.name)}</div>
          </div>
        </div>
        <div class="repo-card-footer">
          <span class="type-badge ${badgeClass[t] || 'badge-model'}">${t}</span>
          ${r.private ? '<span class="type-badge" style="border-color:rgba(100,116,139,.3);color:var(--muted)">private</span>' : ''}
          <span class="repo-date">${date}</span>
        </div>
      </a>`;
  }

  esc(s) {
    return (s || '').replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;');
  }

  showModal() {
    document.getElementById('repoModal').classList.add('show');
  }

  async createRepo() {
    const form = document.getElementById('createRepoForm');
    const fd = new FormData(form);
    const namespace = fd.get('namespace');
    const name = fd.get('name');
    if (!namespace || !name) return;

    const data = {
      name, namespace,
      type: fd.get('type') || 'model',
      private: fd.get('private') === 'on',
      repo_id: `${namespace}/${name}`
    };

    try {
      const res = await fetch('/api/repos/create', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...(this.token ? { Authorization: `Bearer ${this.token}` } : {})
        },
        body: JSON.stringify(data)
      });
      if (!res.ok) throw new Error();
      document.getElementById('repoModal').classList.remove('show');
      form.reset();
      await this.loadRepositories();
      this.showToast('Repository created', 'success');
    } catch {
      this.showToast('Failed to create repository', 'error');
    }
  }

  async login() {
    if (this.authMethods.hf) {
      window.location.href = '/api/auth/hf/login';
      return;
    }
    const token = prompt('Enter your authentication token:');
    if (!token) return;
    try {
      const res = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ token })
      });
      if (!res.ok) throw new Error();
      const data = await res.json();
      this.token = data.token;
      localStorage.setItem('auth_token', this.token);
      this.updateAuthUI();
      this.showToast('Logged in', 'success');
    } catch {
      this.showToast('Login failed', 'error');
    }
  }

  logout() {
    this.token = null;
    localStorage.removeItem('auth_token');
    this.updateAuthUI();
  }

  updateAuthUI() {
    const btn = document.getElementById('loginBtn');
    if (this.token) {
      btn.textContent = 'Logout';
      btn.classList.add('active');
    } else {
      btn.textContent = 'Login';
      btn.classList.remove('active');
    }
  }

  showToast(msg, type = '') {
    const t = document.getElementById('toast');
    t.textContent = msg;
    t.className = 'toast show' + (type ? ` ${type}` : '');
    clearTimeout(this._toastTimer);
    this._toastTimer = setTimeout(() => t.classList.remove('show'), 3000);
  }
}

document.addEventListener('DOMContentLoaded', () => new App());
