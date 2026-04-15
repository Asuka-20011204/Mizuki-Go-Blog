// 封装认证请求
async function authFetch(url: string, options: RequestInit = {}) {
    const token = localStorage.getItem('mizuki_token');
    if (!token) {
        window.location.href = '/admin/login/';
        throw new Error("No token");
    }
    const headers = {
        ...options.headers,
        'Authorization': `Bearer ${token}`
    };
    const response = await fetch(url, { ...options, headers });
    if (response.status === 401) {
        localStorage.removeItem('mizuki_token');
        window.location.href = '/admin/login/';
        throw new Error("Unauthorized");
    }
    return response;
}

// --- 全局函数挂载（供 HTML onclick 使用）---
window.handleDelete = async (slug: string) => {
    if (!confirm(`确定要彻底删除文章 [${slug}] 吗？\n这将同时删除数据库记录和磁盘文件！`)) return;
    try {
        const res = await authFetch(`http://localhost:8080/api/admin/posts/${slug}`, { method: 'DELETE' });
        if (res.ok) {
            // 1. 立即从 DOM 中移除该文章行
            const postRow = document.querySelector(`tr[data-post-slug="${slug}"]`);
            if (postRow) postRow.remove();

            // 2. 提示用户
            alert("文章已物理删除，正在触发重构以清理缓存...");

            // 3. 调用重构函数（复用你已有的重构按钮逻辑）
            await triggerRebuild();  // 需要定义 triggerRebuild 函数
        } else {
            const err = await res.json();
            alert("删除失败：" + (err.error || "未知错误"));
        }
    } catch(e) {
        alert("请求失败");
    }
};

window.deleteDiary = async (id: number) => {
    if (!confirm("确定删除这条日记吗？")) return;
    try {
        const res = await authFetch(`http://localhost:8080/api/admin/diaries/${id}`, { method: 'DELETE' });
        if (res.ok) refreshDiaryList();
    } catch (e) {
        alert("删除失败");
    }
};

window.deleteAlbum = async (id: string) => {
    if (!confirm(`确定要物理删除相册 [${id}] 吗？此操作不可恢复！`)) return;
    try {
        const res = await authFetch(`http://localhost:8080/api/admin/albums/${id}`, { method: 'DELETE' });
        if (res.ok) refreshAlbumList();
    } catch (e) { alert("删除失败"); }
};

window.managePhotos = (albumID: string) => {
    const dialog = document.getElementById('photo-manager-dialog') as any;
    const mgrIdLabel = document.getElementById('mgr-album-id');
    if (!dialog || !mgrIdLabel) return;
    mgrIdLabel.innerText = albumID;
    dialog.showModal();
    refreshPhotoGrid(albumID);
};

window.openPhotoManager = window.managePhotos; // 别名

window.setAsCover = async (albumID: string, filename: string) => {
    const res = await authFetch(`http://localhost:8080/api/admin/albums/${albumID}/set-cover`, {
        method: 'POST',
        body: JSON.stringify({ filename })
    });
    if (res.ok) {
        refreshPhotoGrid(albumID);
        refreshAlbumList();
    }
};

window.deletePhoto = async (albumID: string, filename: string) => {
    if (!confirm(`确定删除图片 ${filename} 吗？`)) return;
    const res = await authFetch(`http://localhost:8080/api/admin/albums/${albumID}/files/${filename}`, {
        method: 'DELETE'
    });
    if (res.ok) refreshPhotoGrid(albumID);
};

// --- 内部数据变量 ---
let diaryUploadedImages: string[] = [];

// --- 刷新各个视图的函数 ---
async function refreshPostList() {
    const tableBody = document.getElementById('posts-list-table');
    if (!tableBody) return;
    try {
        const res = await authFetch('http://localhost:8080/api/admin/posts');
        if (res.status === 401) {
            alert("登录已过期，请重新登录");
            window.location.href = '/admin/login/';
            return;
        }
        const posts = await res.json();
        tableBody.innerHTML = posts.map((post: any) => `
            <tr data-post-slug="${post.Slug || post.slug}" class="border-b dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800/50 transition">
                <td class="p-4 font-medium">${post.Title || post.title}</td>
                <td class="p-4">
                    <span class="px-2 py-1 bg-gray-100 dark:bg-gray-700 rounded text-xs text-gray-600 dark:text-gray-300">
                        ${post.Category || post.category || '未分类'}
                    </span>
                </td>
                <td class="p-4 text-sm text-gray-500">
                    ${new Date(post.CreatedAt || post.created_at).toLocaleDateString()}
                </td>
                <td class="p-4 text-right">
                    <button onclick="location.href='/admin/edit-post/?slug=${post.Slug || post.slug}'" 
                        class="text-blue-500 hover:text-blue-600 font-medium px-2">修改</button>
                    <button onclick="window.handleDelete('${post.Slug || post.slug}')" 
                            class="text-red-500 hover:text-red-600 font-medium px-2">删除</button>
                </td>
            </tr>
        `).join('');
        const postStat = document.getElementById('stat-posts');
        if (postStat) postStat.innerText = posts.length;
    } catch (e) {
        tableBody.innerHTML = '<tr><td colspan="4" class="p-4 text-center text-red-400">无法连接后端，请检查 Go 服务是否启动</td></tr>';
    }
}

async function refreshDiaryList() {
    const diaryList = document.getElementById('diary-list-container');
    if (!diaryList) return;
    try {
        const res = await authFetch('http://localhost:8080/api/admin/diaries');
        const diaries = await res.json();
        diaryList.innerHTML = diaries.map((item: any) => `
            <div class="bg-white dark:bg-[#25262b] p-6 rounded-2xl shadow-sm border border-gray-100 dark:border-gray-800">
                <div class="flex justify-between items-start mb-2">
                    <span class="text-xs text-gray-400 font-mono">${new Date(item.date).toLocaleString()}</span>
                    <div class="flex gap-2">
                        <span class="text-xs px-2 py-1 bg-gray-100 dark:bg-gray-800 rounded text-gray-500">${item.mood || '😶'}</span>
                        <button onclick="window.deleteDiary(${item.id})" class="text-red-400 hover:text-red-500 text-xs">删除</button>
                    </div>
                </div>
                <p class="text-gray-700 dark:text-gray-300 whitespace-pre-wrap">${item.content}</p>
                ${item.location ? `<div class="mt-3 text-xs text-blue-400 flex items-center gap-1">📍 ${item.location}</div>` : ''}
            </div>
        `).join('');
    } catch (e) {
        diaryList.innerHTML = '<p class="text-center text-red-400">加载日记失败</p>';
    }
}

async function refreshAlbumList() {
    const listContainer = document.getElementById('albums-list');
    if (!listContainer) return;
    try {
        const res = await authFetch('http://localhost:8080/api/admin/albums');
        const albums = await res.json();
        listContainer.innerHTML = albums.map((album: any) => `
            <div class="bg-white dark:bg-[#25262b] rounded-2xl shadow-sm border border-gray-100 dark:border-gray-800 overflow-hidden group">
                <div class="aspect-[4/3] bg-gray-100 relative overflow-hidden">
                    <img src="${album.cover}" class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500" 
                        onerror="if(!this.dataset.tried){this.dataset.tried=1;this.src='/images/logo.png';}else{this.style.display='none';}" />
                    <div class="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center gap-4">
                        <button onclick="window.managePhotos('${album.id}')" 
                                class="px-3 py-1 bg-white rounded-full text-orange-500 text-xs font-bold hover:scale-110 transition">
                            管理
                        </button>
                        <button onclick="window.deleteAlbum('${album.id}')" 
                                class="px-3 py-1 bg-white rounded-full text-red-500 text-xs font-bold hover:scale-110 transition">
                            删除
                        </button>
                    </div>
                </div>
                <div class="p-4">
                    <h4 class="font-bold text-slate-800 dark:text-gray-100 truncate">${album.info.title || album.id}</h4>
                    <p class="text-xs text-slate-400 mt-1">${album.info.date || '未设置日期'}</p>
                </div>
            </div>
        `).join('');
    } catch (e) {
        console.error("加载相册失败");
    }
}

async function refreshPhotoGrid(albumID: string) {
    const grid = document.getElementById('mgr-photos-grid');
    if (!grid) return;
    try {
        const res = await authFetch(`http://localhost:8080/api/admin/albums/${albumID}/files`);
        if (res.ok) {
            const files = await res.json();
            if (files.length === 0) {
                grid.innerHTML = '<p class="col-span-4 text-center text-gray-400">暂无图片，请上传</p>';
            } else {
                grid.innerHTML = files.map((file: string) => `
                    <div class="relative group aspect-square rounded-xl overflow-hidden border dark:border-gray-800 bg-gray-50">
                        <img src="/images/albums/${albumID}/${file}?t=${Date.now()}" class="w-full h-full object-cover" />
                        <div class="absolute inset-0 bg-black/60 opacity-0 group-hover:opacity-100 transition-opacity flex flex-col items-center justify-center gap-2">
                            ${file === 'cover.jpg' ? 
                                '<span class="bg-orange-500 text-white text-[10px] px-2 py-1 rounded-full">当前封面</span>' : 
                                `<button onclick="window.setAsCover('${albumID}', '${file}')" class="bg-white text-black text-[10px] px-2 py-1 rounded-full hover:bg-orange-500 hover:text-white transition">设为封面</button>`
                            }
                            <button onclick="window.deletePhoto('${albumID}', '${file}')" class="bg-red-500 text-white text-[10px] px-2 py-1 rounded-full">删除图片</button>
                        </div>
                    </div>
                `).join('');
            }
        } else {
            grid.innerHTML = '<p class="col-span-4 text-center text-red-400">加载失败</p>';
        }
    } catch (e) {
        grid.innerHTML = '<p class="col-span-full text-center text-red-400">读取文件失败</p>';
    }
}

async function refreshStats() {
    try {
        const res = await authFetch('http://localhost:8080/api/admin/stats');
        const data = await res.json();
        const statPosts = document.getElementById('stat-posts');
        const statCategories = document.getElementById('stat-categories');
        const statDays = document.getElementById('stat-days');
        const statWords = document.getElementById('stat-words');
        if (statPosts) statPosts.innerText = data.postCount;
        if (statCategories) statCategories.innerText = data.categoryCount;
        if (statDays) statDays.innerText = data.runningDays;
        if (statWords) {
            statWords.innerText = data.totalWords > 10000 
                ? (data.totalWords / 1000).toFixed(1) + 'k' 
                : data.totalWords;
        }
    } catch (e) {
        console.error("统计数据加载失败");
    }
}

function resetDiaryForm() {
    (document.getElementById('diary-content') as HTMLTextAreaElement).value = "";
    (document.getElementById('diary-location') as HTMLInputElement).value = "";
    (document.getElementById('diary-mood') as HTMLInputElement).value = "";
    diaryUploadedImages = [];
    const previewContainer = document.getElementById('diary-image-preview');
    if (previewContainer) previewContainer.innerHTML = "";
    const diaryForm = document.getElementById('diary-form-container');
    diaryForm?.classList.add('hidden');
}

async function triggerRebuild() {
    const rebuildBtn = document.getElementById('rebuild-btn');
    if (!rebuildBtn) return;
    
    const originalText = rebuildBtn.innerHTML;
    rebuildBtn.disabled = true;
    rebuildBtn.innerHTML = `<span>构建中...</span>`;
    rebuildBtn.classList.replace('bg-green-500', 'bg-gray-500');
    
    try {
        const res = await authFetch('http://localhost:8080/api/admin/rebuild', { method: 'POST' });
        if (res.ok) {
            alert("项目重构成功！新静态内容已同步。");
            // 可选：刷新页面以确保所有缓存都更新
            // window.location.reload();
        } else {
            const err = await res.json();
            alert("重构失败: " + err.error);
        }
    } catch (e) {
        alert("无法连接服务器");
    } finally {
        rebuildBtn.disabled = false;
        rebuildBtn.innerHTML = originalText;
        rebuildBtn.classList.replace('bg-gray-500', 'bg-green-500');
    }
}

// --- 事件绑定和初始化（必须等待 DOM 就绪）---
document.addEventListener('DOMContentLoaded', () => {
    // 检查 token
    const token = localStorage.getItem('mizuki_token');
    if (!token) {
        window.location.href = '/admin/login/';
        return;
    }

    // 侧边栏 Tab 切换
    const navItems = document.querySelectorAll('.nav-item');
    const tabContents = document.querySelectorAll('.tab-content');
    const titleHeader = document.getElementById('current-tab-title');
    navItems.forEach(item => {
        item.addEventListener('click', () => {
            const tabId = item.getAttribute('data-tab');
            navItems.forEach(nav => nav.classList.remove('active'));
            item.classList.add('active');
            tabContents.forEach(content => content.classList.add('hidden'));
            const targetView = document.getElementById(`${tabId}-view`);
            if (targetView) targetView.classList.remove('hidden');
            if (titleHeader) titleHeader.innerText = (item as HTMLElement).innerText.trim();
            if (tabId === 'posts') refreshPostList();
            if (tabId === 'diaries') refreshDiaryList();
            if (tabId === 'dashboard') refreshStats();
            if (tabId === 'albums') refreshAlbumList();
        });
    });

    // 日记相关事件
    const addDiaryBtn = document.getElementById('add-diary-btn');
    const cancelDiaryBtn = document.getElementById('cancel-diary-btn');
    const submitDiaryBtn = document.getElementById('submit-diary-btn');
    const diaryUploadTrigger = document.getElementById('diary-upload-trigger');
    const diaryFileInput = document.getElementById('diary-file-input');
    const photoUploadInput = document.getElementById('photo-upload-input');

    addDiaryBtn?.addEventListener('click', () => {
        document.getElementById('diary-form-container')?.classList.toggle('hidden');
    });
    cancelDiaryBtn?.addEventListener('click', () => {
        if (diaryUploadedImages.length > 0) {
            if (!confirm("已上传的图片将不会被保存，确定取消吗？")) return;
        }
        resetDiaryForm();
    });
    submitDiaryBtn?.addEventListener('click', async () => {
        const content = (document.getElementById('diary-content') as HTMLTextAreaElement).value;
        const location = (document.getElementById('diary-location') as HTMLInputElement).value;
        const mood = (document.getElementById('diary-mood') as HTMLInputElement).value;
        if (!content) return alert("内容不能为空");
        try {
            const res = await authFetch('http://localhost:8080/api/admin/diaries', {
                method: 'POST',
                body: JSON.stringify({ content, location, mood, images: diaryUploadedImages })
            });
            if (res.ok) {
                resetDiaryForm();
                refreshDiaryList();
                alert("动态发布成功！请手动触发重构以更新前台。");
            }
        } catch (e) {
            alert("发布失败");
        }
    });
    diaryUploadTrigger?.addEventListener('click', () => {
        diaryFileInput?.click();
    });
    diaryFileInput?.addEventListener('change', async (e: Event) => {
        const files = (e.target as HTMLInputElement).files;
        if (!files) return;
        const previewContainer = document.getElementById('diary-image-preview');
        for (const file of Array.from(files)) {
            const formData = new FormData();
            formData.append('file', file);
            formData.append('slug', 'diary');
            try {
                const res = await authFetch('http://localhost:8080/api/admin/upload', { method: 'POST', body: formData });
                const data = await res.json();
                if (data.code === 0) {
                    const url = Object.values(data.data.succMap)[0] as string;
                    const freshUrl = `${url}?t=${Date.now()}`;
                    diaryUploadedImages.push(url);
                    const img = document.createElement('img');
                    img.src = freshUrl;
                    img.className = "w-20 h-20 object-cover rounded-lg border dark:border-gray-700";
                    previewContainer?.appendChild(img);
                }
            } catch (err) {
                console.error("图片上传失败", err);
            }
        }
    });

    document.getElementById('dashboard-add-diary')?.addEventListener('click', () => {
        const diaryTabBtn = document.querySelector('.nav-item[data-tab="diaries"]') as HTMLElement;
        diaryTabBtn?.click();
        document.getElementById('add-diary-btn')?.click();
    });

    document.getElementById('dashboard-add-photo')?.addEventListener('click', () => {
        const albumTabBtn = document.querySelector('.nav-item[data-tab="albums"]') as HTMLElement;
        albumTabBtn?.click();
        document.getElementById('new-album-btn')?.click();
    });

    // 相册管理对话框上传
    photoUploadInput?.addEventListener('change', async (e: Event) => {
        const files = (e.target as HTMLInputElement).files;
        const albumID = document.getElementById('mgr-album-id')?.innerText;
        if (!files || !albumID) return;
        for (const f of Array.from(files)) {
            const formData = new FormData();
            formData.append('file', f);
            formData.append('slug', `albums/${albumID}`);
            try {
                const res = await authFetch('http://localhost:8080/api/admin/upload', { method: 'POST', body: formData });
                if (res.ok) console.log(`文件 ${f.name} 上传成功`);
            } catch (err) {
                console.error("上传失败", err);
            }
        }
        alert("上传成功！由于 Astro 静态资源限制，新图片可能需要 Rebuild 后在预览中可见。");
        refreshPhotoGrid(albumID);
        refreshAlbumList();
    });

    // 新建相册对话框
    const newAlbumBtn = document.getElementById('new-album-btn');
    const albumDialog = document.getElementById('album-dialog');
    const confirmCreateAlbum = document.getElementById('confirm-create-album');
    newAlbumBtn?.addEventListener('click', () => {
        (albumDialog as any)?.showModal();
    });
    confirmCreateAlbum?.addEventListener('click', async () => {
        const id = (document.getElementById('album-id') as HTMLInputElement).value.trim();
        const title = (document.getElementById('album-title') as HTMLInputElement).value.trim();
        if (!id || !title) return alert("相册 ID 和标题是必填项");
        const payload = {
            id,
            info: {
                title,
                description: (document.getElementById('album-desc') as HTMLTextAreaElement).value.trim(),
                date: (document.getElementById('album-date') as HTMLInputElement).value || new Date().toISOString().split('T')[0],
                location: (document.getElementById('album-location') as HTMLInputElement).value.trim(),
                tags: (document.getElementById('album-tags') as HTMLInputElement).value.split(',').map(t => t.trim()).filter(Boolean),
                layout: (document.getElementById('album-layout') as HTMLSelectElement).value,
                columns: parseInt((document.getElementById('album-columns') as HTMLInputElement).value) || 3,
                hidden: (document.getElementById('album-hidden') as HTMLInputElement).checked,
                mode: "local"
            }
        };
        try {
            const res = await authFetch('http://localhost:8080/api/admin/albums', { method: 'POST', body: JSON.stringify(payload) });
            if (res.ok) {
                alert("相册配置已生成！");
                (albumDialog as any).close();
                const fields = ['album-id', 'album-title', 'album-desc', 'album-location', 'album-tags'];
                fields.forEach(f => (document.getElementById(f) as any).value = "");
                refreshAlbumList();
            }
        } catch (e) {
            alert("创建失败，请检查后端终端报错");
        }
    });

    // 重构按钮
    const rebuildBtn = document.getElementById('rebuild-btn') as HTMLButtonElement;
    rebuildBtn?.addEventListener('click', async () => {
        const originalText = rebuildBtn.innerHTML;
        rebuildBtn.disabled = true;
        rebuildBtn.innerHTML = `<span>构建中...</span>`;
        rebuildBtn.classList.replace('bg-green-500', 'bg-gray-500');
        try {
            const res = await authFetch('http://localhost:8080/api/admin/rebuild', { method: 'POST' });
            if (res.ok) alert("项目重构成功！新静态内容已同步。");
            else {
                const err = await res.json();
                alert("重构失败: " + err.error);
            }
        } catch (e) {
            alert("无法连接服务器");
        } finally {
            rebuildBtn.disabled = false;
            rebuildBtn.innerHTML = originalText;
            rebuildBtn.classList.replace('bg-gray-500', 'bg-green-500');
        }
    });

    // 退出登录
    const logoutBtn = document.getElementById('logout-btn');
    logoutBtn?.addEventListener('click', () => {
        if (confirm('确定要退出管理系统吗？')) {
            localStorage.removeItem('mizuki_token');
            localStorage.removeItem('mizuki_user');
            window.location.href = '/admin/login/';
        }
    });

    // 初始加载当前激活的 tab
    const activeTab = document.querySelector('.nav-item.active')?.getAttribute('data-tab');
    if (activeTab === 'posts') refreshPostList();
    else if (activeTab === 'diaries') refreshDiaryList();
    else if (activeTab === 'dashboard') refreshStats();
    else if (activeTab === 'albums') {
        refreshAlbumList();
        document.getElementById('albums-view')?.classList.remove('hidden');
    } else {
        // 默认激活 dashboard 但未选中时，可手动刷新 stats
        const dashboardView = document.getElementById('dashboard-view');
        if (dashboardView && !dashboardView.classList.contains('hidden')) {
            refreshStats();
        }
    }
});