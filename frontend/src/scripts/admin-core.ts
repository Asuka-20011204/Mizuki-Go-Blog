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
    } catch (e) {
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
            if (tabId === 'settings') {
                // 加载配置数据
                if (typeof window.loadConfig === 'function') {
                    window.loadConfig();
                }
            }
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


    const saveBtn = document.getElementById('save-config-btn');
    saveBtn?.addEventListener('click', saveConfig);

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
    } else if (activeTab === 'settings') {
        loadConfig();
    } else {
        // 默认激活 dashboard 但未选中时，可手动刷新 stats
        const dashboardView = document.getElementById('dashboard-view');
        if (dashboardView && !dashboardView.classList.contains('hidden')) {
            refreshStats();
        }
    }
});

// ---------- 配置管理 ----------

interface FullConfig {
    siteConfig: any;
    fullscreenWallpaperConfig: any;
    navBarConfig: any;
    profileConfig: any;
    licenseConfig: any;
    expressiveCodeConfig: any;
    commentConfig: any;
    announcementConfig: any;
    musicPlayerConfig: any;
    footerConfig: any;
    sidebarLayoutConfig: any;
    sakuraConfig: any;
    pioConfig: any;
}

let currentConfig: FullConfig | null = null;

async function loadConfig() {
    try {
        const res = await authFetch('http://localhost:8080/api/admin/config');
        if (!res.ok) throw new Error('加载配置失败');
        const config = await res.json() as FullConfig;
        currentConfig = config;
        populateConfigForm(config);  // config 类型明确为 FullConfig
    } catch (e) {
        console.error('加载配置出错', e);
        alert('加载配置失败，请检查后端服务');
    }
}

function populateConfigForm(cfg: FullConfig) {
    const s = cfg.siteConfig;
    const fw = cfg.fullscreenWallpaperConfig;
    const nav = cfg.navBarConfig;
    const prof = cfg.profileConfig;
    const music = cfg.musicPlayerConfig;
    const sakura = cfg.sakuraConfig;
    const pio = cfg.pioConfig;
    const sidebar = cfg.sidebarLayoutConfig;
    const comment = cfg.commentConfig;
    const license = cfg.licenseConfig;
    const code = cfg.expressiveCodeConfig;
    const toc = s.toc;
    const font = s.font;

    // ----- 基础配置 -----
    setValue('cfg-site-title', s.title);
    setValue('cfg-site-subtitle', s.subtitle);
    setValue('cfg-site-url', s.siteURL);
    setValue('cfg-site-start-date', s.siteStartDate?.split('T')[0] || '');
    setValue('cfg-theme-hue', s.themeColor?.hue ?? 230);
    setChecked('cfg-theme-fixed', s.themeColor?.fixed ?? false);
    setValue('cfg-timezone', s.timeZone ?? 8);
    setValue('cfg-lang', s.lang ?? 'zh_CN');

    // 个人资料
    setValue('cfg-profile-avatar', prof?.avatar ?? '');
    setValue('cfg-profile-name', prof?.name ?? '');
    setValue('cfg-profile-bio', prof?.bio ?? '');
    setChecked('cfg-profile-typewriter-enable', prof?.typewriter?.enable ?? true);
    setValue('cfg-profile-typewriter-speed', prof?.typewriter?.speed ?? 80);
    // 社交链接：将对象数组格式化为每行一个 JSON 字符串
    const linksJson = (prof?.links || []).map((l: any) => JSON.stringify(l)).join('\n');
    setValue('cfg-profile-links', linksJson);

    // 公告
    const ann = cfg.announcementConfig;
    setChecked('cfg-announce-enable', ann?.enable ?? true); // 注意公告没有全局 enable，通常直接显示；但保留备用
    setValue('cfg-announce-title', ann?.title ?? '');
    setValue('cfg-announce-content', ann?.content ?? '');
    setChecked('cfg-announce-closable', ann?.closable ?? true);
    setChecked('cfg-announce-link-enable', ann?.link?.enable ?? true);
    setValue('cfg-announce-link-text', ann?.link?.text ?? '');
    setValue('cfg-announce-link-url', ann?.link?.url ?? '');
    setChecked('cfg-announce-link-external', ann?.link?.external ?? false);

    // 页脚
    const footer = cfg.footerConfig;
    setChecked('cfg-footer-enable', footer?.enable ?? false);
    setValue('cfg-footer-html', footer?.customHtml ?? '');

    // 特色页面开关
    const fp = s.featurePages || {};
    setChecked('cfg-page-anime', fp.anime ?? true);
    setChecked('cfg-page-diary', fp.diary ?? true);
    setChecked('cfg-page-friends', fp.friends ?? true);
    setChecked('cfg-page-projects', fp.projects ?? false);
    setChecked('cfg-page-skills', fp.skills ?? true);
    setChecked('cfg-page-timeline', fp.timeline ?? true);
    setChecked('cfg-page-albums', fp.albums ?? true);
    setChecked('cfg-page-devices', fp.devices ?? false);

    // 文章列表布局
    setValue('cfg-postlist-default', s.postListLayout?.defaultMode ?? 'list');
    setChecked('cfg-postlist-switch', s.postListLayout?.allowSwitch ?? true);
    setChecked('cfg-tag-newstyle', s.tagStyle?.useNewStyle ?? false);

    // 整体布局
    setValue('cfg-wallpaper-mode', s.wallpaperMode?.defaultMode ?? 'banner');
    setValue('cfg-mode-switch-display', s.wallpaperMode?.showModeSwitchOnMobile ?? 'desktop');

    // 横幅配置
    const banner = s.banner;
    setTextareaLines('cfg-banner-desktop', banner?.src?.desktop);
    setTextareaLines('cfg-banner-mobile', banner?.src?.mobile);
    setValue('cfg-banner-position', banner?.position ?? 'center');
    setChecked('cfg-banner-carousel-enable', banner?.carousel?.enable ?? true);
    setValue('cfg-banner-carousel-interval', banner?.carousel?.interval ?? 1.5);
    setChecked('cfg-banner-waves-enable', banner?.waves?.enable ?? true);
    setChecked('cfg-banner-waves-perf', banner?.waves?.performanceMode ?? false);
    setChecked('cfg-banner-waves-mobiledisable', banner?.waves?.mobileDisable ?? false);
    setChecked('cfg-banner-homeText-enable', banner?.homeText?.enable ?? true);
    setValue('cfg-banner-homeTitle', banner?.homeText?.title ?? '');
    setTextareaLines('cfg-banner-subtitles', banner?.homeText?.subtitle);
    setChecked('cfg-banner-typewriter-enable', banner?.homeText?.typewriter?.enable ?? true);
    setValue('cfg-banner-typewriter-speed', banner?.homeText?.typewriter?.speed ?? 100);
    setValue('cfg-banner-typewriter-delete', banner?.homeText?.typewriter?.deleteSpeed ?? 50);
    setValue('cfg-banner-typewriter-pause', banner?.homeText?.typewriter?.pauseTime ?? 2000);
    setValue('cfg-navbar-transparent', banner?.navbar?.transparentMode ?? 'semifull');

    // 全屏壁纸
    setTextareaLines('cfg-fullscreen-desktop', fw?.src?.desktop);
    setTextareaLines('cfg-fullscreen-mobile', fw?.src?.mobile);
    setValue('cfg-fullscreen-position', fw?.position ?? 'center');
    setChecked('cfg-fullscreen-carousel-enable', fw?.carousel?.enable ?? true);
    setValue('cfg-fullscreen-carousel-interval', fw?.carousel?.interval ?? 5);
    setValue('cfg-fullscreen-opacity', fw?.opacity ?? 0.8);
    setValue('cfg-fullscreen-blur', fw?.blur ?? 1);
    setValue('cfg-fullscreen-zindex', fw?.zIndex ?? -1);

    // 导航栏标题
    setValue('cfg-navbar-title-text', s.navbarTitle?.text ?? '');
    setValue('cfg-navbar-title-icon', s.navbarTitle?.icon ?? '');

    // 字体配置
    setValue('cfg-font-ascii-family', font?.asciiFont?.fontFamily ?? '');
    setValue('cfg-font-ascii-weight', font?.asciiFont?.fontWeight ?? '400');
    setValue('cfg-font-ascii-files', (font?.asciiFont?.localFonts || []).join(','));
    setChecked('cfg-font-compress-ascii', font?.asciiFont?.enableCompress ?? true);
    setValue('cfg-font-cjk-family', font?.cjkFont?.fontFamily ?? '');
    setValue('cfg-font-cjk-weight', font?.cjkFont?.fontWeight ?? '500');
    setValue('cfg-font-cjk-files', (font?.cjkFont?.localFonts || []).join(','));
    setChecked('cfg-font-compress-cjk', font?.cjkFont?.enableCompress ?? true);

    // 目录 TOC
    setChecked('cfg-toc-enable', toc?.enable ?? true);
    setValue('cfg-toc-depth', toc?.depth ?? 2);
    setChecked('cfg-toc-jpbadge', toc?.useJapaneseBadge ?? true);

    // 评论
    setChecked('cfg-comment-enable', comment?.enable ?? false);
    setValue('cfg-twikoo-envid', comment?.twikoo?.envId ?? '');
    setValue('cfg-twikoo-lang', comment?.twikoo?.lang ?? 'en');

    // 版权
    setChecked('cfg-license-enable', license?.enable ?? true);
    setValue('cfg-license-name', license?.name ?? '');
    setValue('cfg-license-url', license?.url ?? '');

    // 代码块
    setValue('cfg-code-theme', code?.theme ?? 'github-dark');
    setChecked('cfg-code-hideTransition', code?.hideDuringThemeTransition ?? true);

    // 音乐播放器
    setChecked('cfg-music-enable', music?.enable ?? false);
    setValue('cfg-music-mode', music?.mode ?? 'meting');
    setValue('cfg-meting-api', music?.meting_api ?? '');
    setValue('cfg-music-id', music?.id ?? '');
    setValue('cfg-music-server', music?.server ?? 'netease');
    setValue('cfg-music-type', music?.type ?? 'playlist');

    // 樱花特效
    setChecked('cfg-sakura-enable', sakura?.enable ?? false);
    setValue('cfg-sakura-num', sakura?.sakuraNum ?? 21);
    setValue('cfg-sakura-size-min', sakura?.size?.min ?? 0.5);
    setValue('cfg-sakura-size-max', sakura?.size?.max ?? 1.1);
    setValue('cfg-sakura-opacity-min', sakura?.opacity?.min ?? 0.3);
    setValue('cfg-sakura-opacity-max', sakura?.opacity?.max ?? 0.9);
    setValue('cfg-sakura-horiz-min', sakura?.speed?.horizontal?.min ?? -1.7);
    setValue('cfg-sakura-horiz-max', sakura?.speed?.horizontal?.max ?? -1.2);
    setValue('cfg-sakura-vert-min', sakura?.speed?.vertical?.min ?? 1.5);
    setValue('cfg-sakura-vert-max', sakura?.speed?.vertical?.max ?? 2.2);
    setValue('cfg-sakura-rotation', sakura?.speed?.rotation ?? 0.03);
    setValue('cfg-sakura-fade', sakura?.speed?.fadeSpeed ?? 0.03);
    setValue('cfg-sakura-zindex', sakura?.zIndex ?? 100);

    // Live2D 看板娘
    setChecked('cfg-pio-enable', pio?.enable ?? true);
    setValue('cfg-pio-models', (pio?.models || []).join(','));
    setValue('cfg-pio-position', pio?.position ?? 'left');
    setValue('cfg-pio-width', pio?.width ?? 280);
    setValue('cfg-pio-height', pio?.height ?? 250);
    setValue('cfg-pio-mode', pio?.mode ?? 'draggable');
    setChecked('cfg-pio-hidden-mobile', pio?.hiddenOnMobile ?? true);
    setValue('cfg-pio-welcome', pio?.dialog?.welcome ?? '');
    setTextareaLines('cfg-pio-touch', pio?.dialog?.touch);
    setValue('cfg-pio-home', pio?.dialog?.home ?? '');
    setTextareaLines('cfg-pio-skin', pio?.dialog?.skin);
    setValue('cfg-pio-close', pio?.dialog?.close ?? '');
    setValue('cfg-pio-link', pio?.dialog?.link ?? '');

    // 侧边栏布局
    setValue('cfg-sidebar-position', sidebar?.position ?? 'both');
    renderSidebarComponents(sidebar?.components || []);

    // 其他杂项
    setChecked('cfg-show-lastmod', s.showLastModified ?? true);
    setChecked('cfg-generate-og', s.generateOgImages ?? false);
}

function renderSidebarComponents(components: any[]) {
    const container = document.getElementById('sidebar-components-list');
    if (!container) return;

    // 确保组件顺序显示
    container.innerHTML = components.map(comp => {
        const type = comp.type;
        const enable = comp.enable ?? true;
        const order = comp.order ?? 0;
        const sidebar = comp.sidebar ?? 'left';
        return `
        <div class="flex items-center justify-between p-2 border rounded dark:border-gray-700 bg-white dark:bg-gray-800/30">
            <span class="text-sm font-medium">${type}</span>
            <div class="flex items-center gap-2">
                <label class="text-xs flex items-center gap-1">
                    <input type="checkbox" data-comp="${type}-enable" ${enable ? 'checked' : ''} class="accent-blue-500 w-4 h-4" /> 启用
                </label>
                <input type="number" data-comp="${type}-order" value="${order}" class="w-16 p-1 text-sm rounded border dark:bg-gray-800" min="0" />
                <select data-comp="${type}-sidebar" class="text-sm p-1 rounded border dark:bg-gray-800">
                    <option value="left" ${sidebar === 'left' ? 'selected' : ''}>左侧</option>
                    <option value="right" ${sidebar === 'right' ? 'selected' : ''}>右侧</option>
                </select>
            </div>
        </div>
    `}).join('');
}

function collectConfigData(): FullConfig {
    // 构建完整的配置对象，与后端 FullConfig 结构一致
    const s: any = {};
    const fw: any = {};
    const pio: any = {};
    const music: any = {};
    const sakura: any = {};
    const comment: any = { twikoo: {} };
    const license: any = {};
    const code: any = {};

    // 基础
    s.title = getValue('cfg-site-title');
    s.subtitle = getValue('cfg-site-subtitle');
    s.siteURL = getValue('cfg-site-url');
    s.siteStartDate = getValue('cfg-site-start-date');
    s.timeZone = parseInt(getValue('cfg-timezone')) || 8;
    s.lang = getValue('cfg-lang');
    s.themeColor = { hue: parseInt(getValue('cfg-theme-hue')) || 230, fixed: getChecked('cfg-theme-fixed') };

    // 特色页面
    s.featurePages = {
        anime: getChecked('cfg-page-anime'),
        diary: getChecked('cfg-page-diary'),
        friends: getChecked('cfg-page-friends'),
        projects: getChecked('cfg-page-projects'),
        skills: getChecked('cfg-page-skills'),
        timeline: getChecked('cfg-page-timeline'),
        albums: getChecked('cfg-page-albums'),
        devices: getChecked('cfg-page-devices'),
    };

    s.postListLayout = { defaultMode: getValue('cfg-postlist-default'), allowSwitch: getChecked('cfg-postlist-switch') };
    s.tagStyle = { useNewStyle: getChecked('cfg-tag-newstyle') };
    s.wallpaperMode = { defaultMode: getValue('cfg-wallpaper-mode'), showModeSwitchOnMobile: getValue('cfg-mode-switch-display') };

    // 横幅
    s.banner = {
        src: {
            desktop: getTextareaLines('cfg-banner-desktop'),
            mobile: getTextareaLines('cfg-banner-mobile'),
        },
        position: getValue('cfg-banner-position'),
        carousel: {
            enable: getChecked('cfg-banner-carousel-enable'),
            interval: parseFloat(getValue('cfg-banner-carousel-interval')) || 1.5,
        },
        waves: {
            enable: getChecked('cfg-banner-waves-enable'),
            performanceMode: getChecked('cfg-banner-waves-perf'),
            mobileDisable: getChecked('cfg-banner-waves-mobiledisable'),
        },
        homeText: {
            enable: getChecked('cfg-banner-homeText-enable'),
            title: getValue('cfg-banner-homeTitle'),
            subtitle: getTextareaLines('cfg-banner-subtitles'),
            typewriter: {
                enable: getChecked('cfg-banner-typewriter-enable'),
                speed: parseInt(getValue('cfg-banner-typewriter-speed')) || 100,
                deleteSpeed: parseInt(getValue('cfg-banner-typewriter-delete')) || 50,
                pauseTime: parseInt(getValue('cfg-banner-typewriter-pause')) || 2000,
            },
        },
        navbar: { transparentMode: getValue('cfg-navbar-transparent') },
        // 保留原样未修改的字段从 currentConfig 合并，避免丢失
        imageApi: currentConfig?.siteConfig?.banner?.imageApi || { enable: false, url: '' },
        credit: currentConfig?.siteConfig?.banner?.credit || { enable: false, text: '', url: '' },
    };

    // 全屏壁纸
    fw.src = {
        desktop: getTextareaLines('cfg-fullscreen-desktop'),
        mobile: getTextareaLines('cfg-fullscreen-mobile'),
    };
    fw.position = getValue('cfg-fullscreen-position');
    fw.carousel = {
        enable: getChecked('cfg-fullscreen-carousel-enable'),
        interval: parseFloat(getValue('cfg-fullscreen-carousel-interval')) || 5,
    };
    fw.opacity = parseFloat(getValue('cfg-fullscreen-opacity')) || 0.8;
    fw.blur = parseInt(getValue('cfg-fullscreen-blur')) || 1;
    fw.zIndex = parseInt(getValue('cfg-fullscreen-zindex')) || -1;

    // 导航栏标题
    s.navbarTitle = {
        text: getValue('cfg-navbar-title-text'),
        icon: getValue('cfg-navbar-title-icon'),
    };

    // 字体
    s.font = {
        asciiFont: {
            fontFamily: getValue('cfg-font-ascii-family'),
            fontWeight: getValue('cfg-font-ascii-weight'),
            localFonts: getValue('cfg-font-ascii-files').split(',').map(s => s.trim()).filter(Boolean),
            enableCompress: getChecked('cfg-font-compress-ascii'),
        },
        cjkFont: {
            fontFamily: getValue('cfg-font-cjk-family'),
            fontWeight: getValue('cfg-font-cjk-weight'),
            localFonts: getValue('cfg-font-cjk-files').split(',').map(s => s.trim()).filter(Boolean),
            enableCompress: getChecked('cfg-font-compress-cjk'),
        },
    };

    s.toc = {
        enable: getChecked('cfg-toc-enable'),
        depth: parseInt(getValue('cfg-toc-depth')) || 2,
        useJapaneseBadge: getChecked('cfg-toc-jpbadge'),
    };

    s.showLastModified = getChecked('cfg-show-lastmod');
    s.generateOgImages = getChecked('cfg-generate-og');

    // 评论
    comment.enable = getChecked('cfg-comment-enable');
    comment.twikoo = {
        envId: getValue('cfg-twikoo-envid'),
        lang: getValue('cfg-twikoo-lang'),
    };

    // 版权
    license.enable = getChecked('cfg-license-enable');
    license.name = getValue('cfg-license-name');
    license.url = getValue('cfg-license-url');

    // 代码块
    code.theme = getValue('cfg-code-theme');
    code.hideDuringThemeTransition = getChecked('cfg-code-hideTransition');

    // 音乐
    music.enable = getChecked('cfg-music-enable');
    music.mode = getValue('cfg-music-mode');
    music.meting_api = getValue('cfg-meting-api');
    music.id = getValue('cfg-music-id');
    music.server = getValue('cfg-music-server');
    music.type = getValue('cfg-music-type');

    // 樱花
    sakura.enable = getChecked('cfg-sakura-enable');
    sakura.sakuraNum = parseInt(getValue('cfg-sakura-num')) || 21;
    sakura.size = { min: parseFloat(getValue('cfg-sakura-size-min')) || 0.5, max: parseFloat(getValue('cfg-sakura-size-max')) || 1.1 };
    sakura.opacity = { min: parseFloat(getValue('cfg-sakura-opacity-min')) || 0.3, max: parseFloat(getValue('cfg-sakura-opacity-max')) || 0.9 };
    sakura.speed = {
        horizontal: { min: parseFloat(getValue('cfg-sakura-horiz-min')) || -1.7, max: parseFloat(getValue('cfg-sakura-horiz-max')) || -1.2 },
        vertical: { min: parseFloat(getValue('cfg-sakura-vert-min')) || 1.5, max: parseFloat(getValue('cfg-sakura-vert-max')) || 2.2 },
        rotation: parseFloat(getValue('cfg-sakura-rotation')) || 0.03,
        fadeSpeed: parseFloat(getValue('cfg-sakura-fade')) || 0.03,
    };
    sakura.zIndex = parseInt(getValue('cfg-sakura-zindex')) || 100;

    // Pio
    pio.enable = getChecked('cfg-pio-enable');
    pio.models = getValue('cfg-pio-models').split(',').map(s => s.trim()).filter(Boolean);
    pio.position = getValue('cfg-pio-position');
    pio.width = parseInt(getValue('cfg-pio-width')) || 280;
    pio.height = parseInt(getValue('cfg-pio-height')) || 250;
    pio.mode = getValue('cfg-pio-mode');
    pio.hiddenOnMobile = getChecked('cfg-pio-hidden-mobile');
    pio.dialog = {
        welcome: getValue('cfg-pio-welcome'),
        touch: getTextareaLines('cfg-pio-touch'),
        home: getValue('cfg-pio-home'),
        skin: getTextareaLines('cfg-pio-skin'),
        close: getValue('cfg-pio-close'),
        link: getValue('cfg-pio-link'),
    };

    // 侧边栏配置 - 基于原有配置更新可编辑字段
    const originalComponents = currentConfig?.sidebarLayoutConfig?.components || [];
    const updatedComponents = originalComponents.map(comp => {
        const type = comp.type;
        // 从 DOM 获取用户修改的值
        const enableEl = document.querySelector(`[data-comp="${type}-enable"]`) as HTMLInputElement;
        const orderEl = document.querySelector(`[data-comp="${type}-order"]`) as HTMLInputElement;
        const sidebarEl = document.querySelector(`[data-comp="${type}-sidebar"]`) as HTMLSelectElement;

        return {
            ...comp, // 保留原有所有字段
            enable: enableEl ? enableEl.checked : comp.enable,
            order: orderEl ? parseInt(orderEl.value) || 0 : comp.order,
            sidebar: sidebarEl ? sidebarEl.value : comp.sidebar,
        };
    });

    const sidebar = {
        position: getValue('cfg-sidebar-position'),
        components: updatedComponents,
        // 保留原有其他顶层字段
        defaultAnimation: currentConfig?.sidebarLayoutConfig?.defaultAnimation || { enable: true, baseDelay: 0, increment: 50 },
        responsive: currentConfig?.sidebarLayoutConfig?.responsive || {
            breakpoints: { mobile: 768, tablet: 1280, desktop: 1280 },
            layout: { mobile: 'sidebar', tablet: 'sidebar', desktop: 'sidebar' }
        }
    };

    // 个人资料
    const profileLinksText = getValue('cfg-profile-links');
    let profileLinks = [];
    try {
        profileLinks = profileLinksText.split('\n').filter(line => line.trim()).map(line => JSON.parse(line));
    } catch (e) { console.warn('解析社交链接 JSON 失败', e); }
    const profile = {
        avatar: getValue('cfg-profile-avatar'),
        name: getValue('cfg-profile-name'),
        bio: getValue('cfg-profile-bio'),
        typewriter: {
            enable: getChecked('cfg-profile-typewriter-enable'),
            speed: parseInt(getValue('cfg-profile-typewriter-speed')) || 80,
        },
        links: profileLinks,
    };

    // 公告
    const announcement = {
        title: getValue('cfg-announce-title'),
        content: getValue('cfg-announce-content'),
        closable: getChecked('cfg-announce-closable'),
        link: {
            enable: getChecked('cfg-announce-link-enable'),
            text: getValue('cfg-announce-link-text'),
            url: getValue('cfg-announce-link-url'),
            external: getChecked('cfg-announce-link-external'),
        },
    };

    // 页脚
    const footer = {
        enable: getChecked('cfg-footer-enable'),
        customHtml: getValue('cfg-footer-html'),
    };
    // 在 return 语句中合并
return {
    siteConfig: { ...currentConfig?.siteConfig, ...s },
    fullscreenWallpaperConfig: { ...currentConfig?.fullscreenWallpaperConfig, ...fw },
    navBarConfig: currentConfig?.navBarConfig || {},
    profileConfig: { ...currentConfig?.profileConfig, ...profile },
    licenseConfig: { ...currentConfig?.licenseConfig, ...license },
    expressiveCodeConfig: { ...currentConfig?.expressiveCodeConfig, ...code },
    commentConfig: { ...currentConfig?.commentConfig, ...comment },
    announcementConfig: { ...currentConfig?.announcementConfig, ...announcement },
    musicPlayerConfig: { ...currentConfig?.musicPlayerConfig, ...music },
    footerConfig: { ...currentConfig?.footerConfig, ...footer },
    sidebarLayoutConfig: { ...currentConfig?.sidebarLayoutConfig, ...sidebar },
    sakuraConfig: { ...currentConfig?.sakuraConfig, ...sakura },
    pioConfig: { ...currentConfig?.pioConfig, ...pio },
};
}

async function saveConfig() {
    if (!currentConfig) {
        alert('请先加载配置');
        return;
    }
    const payload = collectConfigData();
    const statusEl = document.getElementById('config-save-status');
    try {
        const res = await authFetch('http://localhost:8080/api/admin/config', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });
        if (res.ok) {
            if (statusEl) {
                statusEl.classList.remove('hidden');
                statusEl.innerText = '✅ 配置已保存，请触发重构使更改生效';
                setTimeout(() => statusEl.classList.add('hidden'), 4000);
            }
            // 更新本地缓存
            currentConfig = payload;
        } else {
            const err = await res.json();
            alert('保存失败: ' + (err.error || '未知错误'));
        }
    } catch (e) {
        alert('网络错误');
    }
}

// 辅助函数
function setValue(id: string, value: any) {
    const el = document.getElementById(id) as HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement;
    if (el) el.value = value?.toString() ?? '';
}
function setChecked(id: string, checked: boolean) {
    const el = document.getElementById(id) as HTMLInputElement;
    if (el) el.checked = checked;
}
function getValue(id: string): string {
    const el = document.getElementById(id) as any;
    return el?.value ?? '';
}
function getChecked(id: string): boolean {
    const el = document.getElementById(id) as HTMLInputElement;
    return el?.checked ?? false;
}
function setTextareaLines(id: string, arr: string[] | undefined) {
    const el = document.getElementById(id) as HTMLTextAreaElement;
    if (el && Array.isArray(arr)) el.value = arr.join('\n');
    else if (el) el.value = '';
}
function getTextareaLines(id: string): string[] {
    const el = document.getElementById(id) as HTMLTextAreaElement;
    return el?.value.split('\n').map(s => s.trim()).filter(Boolean) || [];
}

// 挂载到 window
window.loadConfig = loadConfig;
window.saveConfig = saveConfig;