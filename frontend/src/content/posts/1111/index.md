---
title: 为 Mizuki 博客后台新增管理员用户管理及移动端适配全记录
published: 2026-05-07
description: 记录为 Mizuki 博客后台新增管理员协作功能与侧边栏移动端适配，包含前后端完整实现与排障过程。
image: "/images/albums/AcgExample/4.webp"
tags:
  - Go
  - Mizuki
  - Astro
  - TypeScript
  - 教程
category: 教程
pinned: false
lang: "zh-CN"
draft: false
---

本文记录了为 Mizuki 博客后台新增两项功能的完整过程：侧边栏移动端适配，以及管理员用户管理（支持多人协作但不开放注册）。同时包含了开发过程中踩到的四个坑和排查思路。

## 一、侧边栏移动端适配

### 1.1 原有问题

原侧边栏组件 `Sidebar.astro` 上只有一句 `hidden md:flex`——屏幕宽度不足 768px 时整条侧边栏完全消失，而且整个页面没有任何菜单按钮。手机上登后台时完全无法导航。

### 1.2 改造思路

不改布局结构，只改 CSS：移动端将侧边栏从"隐藏"变成"侧滑面板"——默认移出屏幕左侧，点击汉堡按钮后滑入，带半透明遮罩。桌面端行为不变。

### 1.3 具体修改

涉及三个文件。

**Sidebar.astro**：新增遮罩 div，aside 加 id 和过渡 class。

```html
<!-- 移动端遮罩 -->
<div id="sidebar-overlay" class="fixed inset-0 bg-black/50 z-30 hidden" aria-hidden="true"></div>

<aside id="admin-sidebar"
  class="fixed md:static inset-y-0 left-0 z-40 w-64 bg-white dark:bg-[#25262b] shadow-xl flex flex-col
         transform -translate-x-full md:translate-x-0 transition-transform duration-300 ease-in-out">
```

关键的 Tailwind class 组合：

- `fixed md:static`：移动端固定定位悬浮于页面之上，大屏恢复到文档流
- `-translate-x-full md:translate-x-0`：移动端默认向左移出屏幕（隐藏），大屏复位
- `transition-transform duration-300`：显隐切换时 300ms 平滑滑动

header 内加关闭按钮，`md:hidden` 只在移动端显示。底部"返回前台"和"退出登录"保留不动。

**index.astro**：顶栏左侧加汉堡按钮，同样 `md:hidden`。

```html
<button id="menu-toggle-btn" class="md:hidden p-2 ...">
  <Icon name="material-symbols:menu" class="text-2xl" />
</button>
```

**admin-core.ts**：在 DOMContentLoaded 回调里加开关逻辑。

```typescript
const sidebar = document.getElementById('admin-sidebar');
const overlay = document.getElementById('sidebar-overlay');
const menuToggle = document.getElementById('menu-toggle-btn');
const sidebarClose = document.getElementById('sidebar-close-btn');

function openSidebar() {
    sidebar?.classList.remove('-translate-x-full');
    sidebar?.classList.add('translate-x-0');
    overlay?.classList.remove('hidden');
    document.body.style.overflow = 'hidden';   // 防止背景滚动
}

function closeSidebar() {
    sidebar?.classList.add('-translate-x-full');
    sidebar?.classList.remove('translate-x-0');
    overlay?.classList.add('hidden');
    document.body.style.overflow = '';
}

menuToggle?.addEventListener('click', openSidebar);
sidebarClose?.addEventListener('click', closeSidebar);
overlay?.addEventListener('click', closeSidebar);

// Escape 键关闭
document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape' && sidebar?.classList.contains('translate-x-0')) {
        closeSidebar();
    }
});
```

同时，在已有的导航切换逻辑末尾加一句 `closeSidebar()`——移动端点完菜单项后自动收起来，不用手动关。

## 二、管理员用户管理

项目的 User 模型和 JWT 认证早就有了，只需要在它的 MVC 分层上补一条完整的用户管理链路。

### 2.1 角色设计

User 表有两个角色：`owner`（所有者，唯一且不可删除）和 `admin`（普通管理员，由 owner 创建）。首次启动时 `InitDatabase` 用 config.yaml 的 init_data 创建 owner。

安全边界就两条：

- 只有 owner 能加人和删人
- 任何人都不能删 owner

### 2.2 后端实现

**Models**——在 `models/user.go` 追加三个结构体：

```go
// 添加管理员的请求体
type AddAdminRequest struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
    QQ       string `json:"qq"`
    Phone    string `json:"phone"`
}

// 修改密码的请求体
type ChangePasswordRequest struct {
    OldPassword string `json:"old_password" binding:"required"`
    NewPassword string `json:"new_password" binding:"required"`
}

// 返回给前端的用户信息（不含密码字段）
type UserInfo struct {
    ID       uint   `json:"id"`
    Username string `json:"username"`
    QQ       string `json:"qq"`
    Phone    string `json:"phone"`
    Avatar   string `json:"avatar"`
    Role     string `json:"role"`
}
```

**Service**——新建 `service/user_service.go`，四个方法：

`ListAdmins`：查全部用户，循环转成 UserInfo 脱敏后返回。

`AddAdmin`：先查用户名是否已被占用，然后用 bcrypt 对密码做哈希。如果提供了 QQ 号，自动拼接 QQ 头像 URL。新用户角色固定为 `admin`，不允许通过接口创建 owner。

`DeleteAdmin`：三重保护——传入的 userID 不能等于当前登录者 ID、用 First 确认用户存在、user.Role 为 `"owner"` 时拒绝。

`ChangePassword`：先 `bcrypt.CompareHashAndPassword` 验证旧密码，通过后再生成新哈希写入。

**Controller**——新建 `controller/user_ctrl.go`，权限校验在这一层完成：

```go
func (uc *UserController) AddAdmin(c *gin.Context) {
    role, _ := c.Get("role")   // JWT 中间件在验证 token 后注入 context
    if role != "owner" {
        c.JSON(http.StatusForbidden, gin.H{"error": "只有所有者才能添加管理员"})
        return
    }
    // ... 绑定 JSON，调 service
}
```

userID 和 role 来自 JWT 中间件 `c.Set("userID", claims.UserID)` 和 `c.Set("role", claims.Role)`，前端无法伪造。

**Router**——在 `main.go` 注册四条路由：

```go
admin.GET("/users", userCtrl.ListAdmins)
admin.POST("/users", userCtrl.AddAdmin)
admin.DELETE("/users/:id", userCtrl.DeleteAdmin)
admin.PUT("/users/password", userCtrl.ChangePassword)
```

### 2.3 前端实现

新增 `UsersView.astro` 组件，模板风格与其他管理视图一致：

- 顶部操作栏（标题 + 修改密码按钮 + 添加管理员按钮）
- 用户列表表格（头像、用户名、角色标签、联系方式、操作）
- 两个 `<dialog>` 对话框（添加管理员、修改密码）

Sidebar 的"系统设置"分组里加一项 `<button class="nav-item" data-tab="users">`。index.astro import 后放在主内容区即可——tab 切换是基于 `data-tab` 属性自动匹配的，加一个视图不需要改切换逻辑本身。

**前端权限判断**——解码 JWT 取 role，不需要额外请求：

```typescript
function getCurrentUserRole(): string {
    const token = localStorage.getItem('mizuki_token');
    if (!token) return '';
    try {
        const payload = JSON.parse(atob(token.split('.')[1]));
        return payload.role || '';
    } catch {
        return '';
    }
}
```

UI 层面的控制：owner 看到"添加管理员"按钮和每行后面的"删除"按钮，admin 看不到。真正的权限强制靠后端 Controller 层，前端只是显示/隐藏。

## 三、踩坑记录

### 3.1 按钮颜色透明看不见

**现象**：添加管理员的按钮和两个对话框的确认按钮完全看不见。

**原因**：按钮用了 `bg-[var(--primary)]`——这是前台主题注入的 CSS 变量，后台管理页面没有加载这套注入逻辑，变量未定义，背景色回退为 transparent。

**解决**：后台按钮不用 CSS 变量，直接写颜色值。实心按钮 `bg-blue-600 hover:bg-blue-700 text-white`；边框按钮额外指定文字色和边框色 `border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-200`。新旧两种按钮都改掉。

### 3.2 API 绝对路径换环境就挂

**现象**：代码里大量 `authFetch('http://localhost:8080/api/admin/xxx')`，本地开发可能没问题，换端口或部署到线上直接全挂。

**解决**：所有 `http://localhost:8080` 前缀删掉，换成 `/api/admin/xxx` 相对路径。浏览器自动拼协议、域名、端口。但光改相对路径还不够，看下面这条。

### 3.3 斜杠冲突（最折腾的问题）

**现象**：改完相对路径后，API 请求仍然 404。在 Vite 里加了 proxy，时好时坏，部分接口能通、部分不通。Network 面板看到 301 Moved Permanently。

**排查**：三层问题叠加。

第一层是 Astro 的 `trailingSlash: 'always'`——收到 `/api/admin/stats`（无尾斜杠）直接拦截并提示"你要不要试试 `/api/admin/stats/`？"，返回 404。

第二层是 Vite proxy 的执行顺序——proxy 配置在 Astro 中间件之后执行，先被 trailingSlash 拦住了。

第三层是 Gin 的 `RedirectTrailingSlash`（默认 true）——给 `/api/admin/stats/`（带斜杠）的请求自动 301 重定向到 `/api/admin/stats`（无斜杠），浏览器跟过去又撞上 Astro 拦截，死循环。

所以完整的事故链是：无斜杠请求 → Astro 拦截 404。加斜杠请求 → Vite proxy 转发成功 → Gin 301 去掉斜杠 → 浏览器跟过去 → 又变无斜杠 → Astro 再次拦截。

浏览器对 301 响应会做磁盘缓存，导致不同接口表现不一致——有些接口之前成功过走了缓存所以还能通，有些不通，增加了排查迷惑性。

**解决**：前后端各改一处。

前端 `authFetch` 保留主动补尾斜杠的逻辑：

```typescript
const finalUrl = url.includes('?') ? url : (url.endsWith('/') ? url : url + '/');
```

后端 `main.go` 关掉 Gin 的重定向并加去尾斜杠中间件：

```go
r := gin.Default()
r.RedirectTrailingSlash = false

// 在 CORS 之前注册，自动去掉 URL 尾部斜杠
r.Use(func(c *gin.Context) {
    path := c.Request.URL.Path
    if len(path) > 1 && strings.HasSuffix(path, "/") {
        c.Request.URL.Path = path[:len(path)-1]
    }
    c.Next()
})
```

再在 `astro.config.mjs` 的 vite 段配置 proxy：

```js
vite: {
    server: {
        proxy: {
            '/api': 'http://localhost:8080',
            '/preview-cache': 'http://localhost:8080',
        },
    },
    // ...
}
```

最终链路：

```text
浏览器 fetch('/api/admin/stats')
  → authFetch 加斜杠 → /api/admin/stats/
  → Astro 放行（已有斜杠）
  → Vite proxy 转发 :8080
  → Go 中间件去斜杠 → /api/admin/stats
  → Gin 直接匹配路由 ✅
```

### 3.4 GORM record not found 日志

**现象**：添加管理员时终端输出 `record not found`，以为报错了，但实际上返回 200 创建成功。

**原因**：`AddAdmin` 用 `First()` 检查用户名是否已存在。用户名不存在时 GORM 会打这条日志，而"不存在"正好是期望的结果。日志级别是 info，不是 error。

**解决**：用 `Count` 替代 `First` 做查重就不会触发这条日志了。

```go
var count int64
s.DB.Model(&models.User{}).Where("username = ?", req.Username).Count(&count)
if count > 0 {
    return errors.New("用户名已被占用")
}
```

## 四、总结

改动文件清单：

后端（Go）：

| 文件                           | 操作                                                                  |
| ------------------------------ | --------------------------------------------------------------------- |
| `models/user.go`               | 追加 3 个 DTO 结构体                                                  |
| `service/user_service.go`      | 新建，4 个业务方法                                                    |
| `controller/user_ctrl.go`      | 新建，4 个接口处理器                                                  |
| `main.go`                      | 依赖注入、路由注册、RedirectTrailingSlash 关闭、去尾斜杠中间件        |

前端（Astro / TypeScript）：

| 文件                                | 操作                                                        |
| ----------------------------------- | ----------------------------------------------------------- |
| `components/admin/Sidebar.astro`    | 移动端侧滑面板、遮罩、关闭按钮、用户管理导航                |
| `components/admin/UsersView.astro`  | 新建，用户列表 + 添加/修改密码对话框                          |
| `pages/admin/index.astro`           | 导入 UsersView、汉堡菜单按钮                                 |
| `scripts/admin-core.ts`             | 移动端菜单开关、slash 适配、用户 CRUD 逻辑                   |
| `astro.config.mjs`                  | Vite proxy 配置                                             |
