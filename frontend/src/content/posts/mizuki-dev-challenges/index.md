---
title: Mizuki 博客后台改造——重难点与踩坑全记录
published: 2026-05-08
description: 记录 Mizuki 博客后台开发中遇到的所有重难点、设计决策、跨框架冲突排查过程，以及对应的解决方案。面向个人复习和项目复盘。
image: "/images/albums/AcgExample/4.webp"
tags:
  - Go
  - Mizuki
  - Astro
  - 复盘
  - 踩坑
category: 技术
pinned: true
lang: "zh-CN"
draft: false
---

这篇文章不是教程，是我对 Mizuki 博客后台这段时间开发改造的一次完整复盘。主要记录遇到的重难点、踩过的坑、以及当时的排查思路和最终方案。以后自己复习或者面试的时候可以直接翻。

---

## 一、文章双写一致性（MySQL + Markdown 文件）

这是整个项目架构上最核心的难点。文章同时存在于两个地方——MySQL 数据库和 Markdown 文件——需要保证两边数据一致。

### 1.1 为什么要双写

文章有两个消费者：

- **管理后台**需要列表查询、搜索、统计，这些 SQL 最擅长。如果只有文件系统，后台文章列表就要遍历目录一个个读 Markdown 再解析，文章多了性能扛不住。
- **Astro 前台**用的是 content collection 机制，要求文章以 Markdown 文件的形式存在于 `src/content/posts/` 目录下，构建时直接读取文件生成静态页面。如果只有数据库，前台就找不到文章。

所以选了双写——每次发布，同时写一份到数据库、一份到文件系统。各取所需。

### 1.2 两边的数据结构差异

数据库的 Post 表是扁平字段：

```go
type Post struct {
    gorm.Model
    Title    string
    Slug     string `gorm:"uniqueIndex;not null"`
    Content  string `gorm:"type:longtext"`  // 纯正文，不含 Frontmatter
    Category string
    Tags     string  // 逗号分隔字符串，如 "Go,Astro"
    Pinned   bool
    // ...
}
```

Markdown 文件是 Frontmatter + 正文的结构：

```markdown
---
title: "我的第一篇文章"
published: 2026-05-08
tags: ['Go', 'Astro']
category: "技术"
pinned: false
---

这里是正文...
```

注意 Tags 字段在数据库里是 `"Go,Astro"`（逗号分隔的字符串，方便 SQL 模糊查询），在 Markdown 里是 `['Go', 'Astro']`（YAML 数组，方便 Astro 识别）。两边格式不同，写入和读取时都需要转换。

### 1.3 写入路径

`ProcessPostPublish` 是整个写入的入口，按顺序做三件事：

**第一步：清洗正文**。编辑器传过来的正文可能不小心把 Frontmatter 也带进来了（用户从别处粘贴了带 frontmatter 的内容）。用 Go 的 `strings.SplitN` 按 `"---"` 把内容切成三段，取最后一段作为纯正文。这样保证写入文件时不会出现双层 `---`。

**第二步：拼 Markdown 写文件**。用 Go 的 `fmt.Sprintf` 把数据库字段按模板格式序列化为完整的 Markdown 字符串，其中 tags 切片用 `strings.Join` 拼成 `['tag1', 'tag2']` 格式，布尔值直接用 `%v` 格式化。然后 `os.WriteFile` 写到 `posts/{slug}/index.md`。

```go
mdTemplate := `---
title: "%s"
published: %s
tags: %s
category: "%s"
pinned: %v
---

%s`

fullContent := fmt.Sprintf(mdTemplate, req.Title, publishedDate,
    formattedTags, req.Category, req.Pinned, cleanBody)
os.WriteFile(filepath.Join(postDir, "index.md"), []byte(fullContent), 0644)
```

**第三步：同步数据库**。用 Slug 做唯一键，GORM 的 `Unscoped().Where("slug = ?").Assign(post).FirstOrCreate(&post)`——Slug 不存在就创建新行，已存在就更新全部字段（包括之前被软删除的文章也会被重新激活）。

**容易出问题的地方**：这三步中，文件先写、数据库后写。如果第三步数据库写入失败，文件已经落盘了但没有数据库记录——这篇文章在后台列表中看不到，但文件已经存在。目前没有在写文件失败时回滚文件的机制（Go 里需要手动处理），实际使用中两边同时出错概率极低，就没上事务。但如果要做生产级系统，这里需要改进。

### 1.4 读取路径

编辑文章时请求 `GET /api/admin/posts/:slug`，`GetPostBySlug` 做了"两个来源拼一份数据"的事情：

先从数据库查出 Post 行（拿到标题、分类、标签等元数据），然后从文件系统读取 `posts/{slug}/index.md`，用 `strings.SplitN` 剥离 Frontmatter 只取正文，覆盖掉 `post.Content` 字段返回。

**为什么从文件拿正文而不是从数据库？** 因为用户可能在服务器上直接编辑了 Markdown 文件。如果从数据库拿，拿到的是发布时的旧快照；从文件拿，拿到的是最新的真实内容。

### 1.5 Frontmatter 的前后端双重解析

这是数据流里隐藏最深的坑。整个项目里，Markdown 的 YAML Frontmatter 在 **Go 端（写入生成）和 JS 端（编辑回显解析）各实现了一套**，没有任何共享的类型约束。

**后端 Go 生成 Frontmatter**：用 `fmt.Sprintf` 硬编码模板。Tags 的输出格式是 `['Go', 'Astro']`（单引号包裹的 JSON 风格）。

**前端 JS 解析 Frontmatter**：`edit-post.astro` 里有个手写的 `parseFrontmatter` 函数（大约 50 行），它是纯浏览器端 YAML 解析器，没用 js-yaml 库。解析逻辑是逐行 split、冒号分割 key/value，然后做类型推断：

```javascript
// 布尔值
if (rawValue === 'true') { data[key] = true; }

// 数组 [a, b, c]
if (rawValue.startsWith('[') && rawValue.endsWith(']')) {
    data[key] = arrContent.split(',').map(v => v.trim().replace(/^["']|["']$/g, ''));
}
```

**问题在哪**：后端生成的格式和前端期望的格式必须完全一致。Tags 数组必须以 `['...', '...']` 格式输出，引号必须是单引号。如果后端某天改了模板、或者字符串里出现了特殊字符导致输出 `[Go, Astro]`（无引号），前端的 split 解析就会出错。两边没有任何编译期检查，全靠约定，改代码时很容易踩。

---

## 二、混合存储策略

我一直觉得这是项目中一个值得说的设计点。不同数据用了不同的存储方案，不是一刀切全部 MySQL 或全部文件。

| 数据 | 存储 | 为什么 |
|------|------|--------|
| 文章 | MySQL + 文件 | 后台需要列表和搜索（SQL），前台构建需要 Markdown 文件 |
| 日记 | JSON 文件 + TS 文件 | 数据量小、结构简单、前台静态引用不需要请求 |
| 相册 | 文件系统 + info.json | 图片天然适合文件系统，元数据 JSON 就够 |
| 用户 | MySQL | 需要唯一索引、关联查询 |
| 站点配置 | JSON 文件 | 单例配置、前台和后台都要读、更新频率低 |

日记如果建 MySQL 表是过度设计——总共几十条记录，字段只有内容、日期、心情、地点、图片。用 JSON 文件读写完全够用，代码也更简单。

但用 JSON 文件有一个问题：**Astro 前台是静态站点，不能发 HTTP 请求来读 JSON 文件**。所以我的方案是——后端每次增删日记后，在写入自己用的 JSON 的同时，把同样的数据序列化成 TypeScript export 语句写入前端源码目录：

```go
tsDataContent := fmt.Sprintf(
    "// 此文件由后端自动生成\nexport const diaryData = %s;",
    string(jsonData),
)
os.WriteFile("../frontend/src/data/diary_data.ts", []byte(tsDataContent), 0644)
```

前台直接 `import { diaryData } from "@/data/diary_data"` 就能拿到编译期的静态数据，零网络请求。代价是每次日记更新后如果要上线，需要 rebuild 重新打包。

---

## 三、图片上传的上下文感知目录路由

上传接口只有一个 `POST /api/admin/upload`，接收 `file` + `slug` 两个参数。后端的 `SavePostResource` 根据 slug 前缀决定图片存在哪里：

- `slug = "my-post"` → 文章图片。先**只存 preview-cache 目录**，发布时才把图片搬到 `posts/my-post/` 下。这样用户上传了图片但放弃发布时，不会在 posts 目录留下垃圾文件夹。
- `slug = "albums/东京"` → 相册图片。存到 `public/images/albums/东京/`，直接可用。
- `slug` 为空 → 兜底存 preview-cache。

单一接口承载三种存储策略，不需要为文章图片和相册图片各写一个上传接口。

### 3.1 图片路径的双向转换

**上传时**：图片存在 preview-cache，Vditor 编辑器插入的 URL 是 `/preview-cache/abc.jpg`。

**发布时**：需要把正文里的 `/preview-cache/abc.jpg` 替换为 `./abc.jpg`（相对路径），同时把对应的图片文件从 preview-cache 复制到文章目录。这两个操作——替换路径和搬运文件——都必须发生，漏一个就裂图。

**编辑回显时**：Markdown 里的 `./abc.jpg` 需要换回 `/preview-cache/abc.jpg`，否则编辑器里图片不显示。

三个上下文、三种路径格式，两个方向的替换逻辑。哪个方向搞错图片就裂。

### 3.2 图片搬运的匹配时机 Bug

这是我开发过程中发现的一个实际问题。前端 `new-post.astro` 在发布前做了一步预处理：

```javascript
const rawContent = window.vditor.getValue();
const finalContent = rawContent.replaceAll('/preview-cache/', './');
```

把 `/preview-cache/` 全部换成 `./`，然后 POST 给后端。但后端的 `ProcessPostPublish` 里搬运图片用的是这个正则：

```go
re := regexp.MustCompile(`/preview-cache/([^\s\)]+)`)
matches := re.FindAllStringSubmatch(req.Content, -1)
```

前端已经把 `/preview-cache/` 换成 `./` 了才发过来，后端再用 `/preview-cache/` 去正则匹配，**永远匹配不到**。图片搬运代码实际上从来没执行过。

**修复方式**：把路径替换从前端移到后端——前端发送原始内容（保留 `/preview-cache/` 路径），后端先匹配搬运图片，搬运完成后再把正文里的路径统一替换为 `./`。

这个 bug 能活下来是因为 preview-cache 是 Astro 的静态目录，即使图片没搬到文章目录，前台通过 `/preview-cache/` 路径依然能访问——但这不是正确架构，等于图片永远留在临时目录里。

### 3.3 Vite 不识别运行时新写入的 public/ 文件

相册图片存在 `frontend/public/images/albums/` 下。`pnpm dev` 启动时，Vite 会缓存 public 目录的文件列表。后端通过 API 写入的新文件 Vite 不知道，请求 `/images/albums/东京/新图.jpg` 时返回空。

**解决**：所有用户运行时上传的资源目录不交给 Astro/Vite 托管，改用 Go 的 `r.Static()` 直接 serve。在 `main.go` 里：

```go
r.Static("/preview-cache", "../frontend/public/preview-cache")
r.Static("/images/albums", "../frontend/public/images/albums")
```

然后在 `astro.config.mjs` 的 Vite proxy 里把这两个路径也转发到 Go 后端。这样开发时 Go 直接读磁盘返回文件，不经过 Vite 缓存；生产时 Nginx 直接 serve 这个目录。

---

## 四、权限体系的前后端双重校验

这是安全设计上比较重要的一环。系统有两个角色——owner（所有者，唯一）和 admin（普通管理员）。

### 4.1 三层防线

**第一层——JWT 中间件**：解析请求头中的 `Authorization: Bearer <token>`，验证签名和过期时间。通过后将 `userID`、`username`、`role` 注入 Gin context。非法或过期 token 直接 401。

**第二层——Controller 权限校验**：对敏感操作（添加管理员、删除管理员）校验 `role == "owner"`，不通过直接 403。这一层是**真正的安全边界**。

```go
func (uc *UserController) AddAdmin(c *gin.Context) {
    role, _ := c.Get("role")   // JWT 中间件注入的上下文，无法伪造
    if role != "owner" {
        c.JSON(http.StatusForbidden, gin.H{"error": "只有所有者才能添加管理员"})
        return
    }
```

**第三层——Service 业务规则**：删除管理员时三重保护——不能删除自己、用 First 确认用户存在、检查目标用户的 Role 是否为 owner。

### 4.2 前端 UI 控制

前端通过解码 JWT payload 获取当前用户的角色，控制按钮的显隐——admin 看不到"添加管理员"和"删除"按钮。

```typescript
function getCurrentUserRole(): string {
    const token = localStorage.getItem('mizuki_token');
    const payload = JSON.parse(atob(token.split('.')[1]));
    return payload.role || '';
}
```

**前端隐藏只是体验优化，不是安全措施**。懂行的人可以手动构造请求绕过前端隐藏。真正的安全靠后端 Controller 层——前端怎么改都没用，后端会校验 role。

如果面试问到这一点，关键是说清楚"权限判断放在哪一层"——放在中间件太粗糙（有些接口 owner 和 admin 都能访问），放在 Service 太晚（参数已经解析了），放在 Controller 刚好。

---

## 五、Astro trailingSlash × Vite Proxy × Gin 的三方死锁

这是整个项目里排查时间最长的问题，跨了三个框架，每个框架各有一个"斜杠策略"，组合在一起形成死锁。写下来以便以后自己回顾。

### 5.1 背景：为什么要改 API 路径

项目原来的代码里所有 API 请求都写了绝对路径 `http://localhost:8080/api/admin/xxx`。这在本地开发时能通（直接打到 Go 后端端口），但部署到服务器上就全挂——前端的 `localhost:8080` 在用户浏览器里没有意义。

所以第一步改动是把所有绝对路径改成相对路径：`http://localhost:8080/api/admin/posts` → `/api/admin/posts`。浏览器会基于当前页面的域名和端口自动拼接完整 URL。

改完之后所有请求 404。排查发现是 Astro 的 `trailingSlash: 'always'` 配置——它要求所有 URL 以 `/` 结尾，`/api/admin/posts` 没有斜杠，直接拦截返回：

```
404: Not found. Did you mean /api/admin/posts/?
```

### 5.2 每个组件做了什么

为了理解这个问题，先梳理一下三个组件各自的斜杠行为：

**Astro trailingSlash: 'always'**

收到无尾斜杠的请求直接返回 404 并提示"要不要加个 /？"。所有路径都要以 `/` 结尾。

**Vite server.proxy**

开发代理，匹配 `/api` 前缀的请求转发到 Go 后端（端口 8080）。本意是让 API 请求绕过 Astro 路由直接到后端。

**Gin RedirectTrailingSlash（默认 true）**

路由注册的是 `/api/admin/posts`（无斜杠），如果请求到达的是 `/api/admin/posts/`（带斜杠），Gin 自动返回 301 重定向到无斜杠版本。

### 5.3 死锁的形成过程

单独看每个组件的行为都没问题，但组合起来：

**第一步**：浏览器发起 `fetch('/api/admin/stats')`，无斜杠。Astro 拦截，返回 404。

**第二步**：为了绕开 Astro，在 `authFetch` 里主动给所有 API 请求补尾斜杠：`/api/admin/stats` → `/api/admin/stats/`。Astro 检查有斜杠，放行。

**第三步**：Vite proxy 匹配到 `/api`，转发到 Go 后端（`http://localhost:8080/api/admin/stats/`）。

**第四步**：Gin 收到带斜杠的请求。路由注册的是 `/stats` 不是 `/stats/`。`RedirectTrailingSlash = true`，Gin 返回 301：`Location: /api/admin/stats`。

**第五步**：浏览器收到 301，跟随重定向。新 URL 是 `/api/admin/stats`——又变回无斜杠。回到第一步。

死锁形成：**无斜杠 → Astro 拦截。加斜杠 → 穿过 Astro 穿过 Vite → Gin 301 去斜杠 → 浏览器重定向 → 又无斜杠 → Astro 拦截**。

### 5.4 为什么有的请求能通有的不能

这是最迷惑人的地方。浏览器对 301 响应做了**磁盘缓存**（Network 面板显示 `(从磁盘缓存)`）。之前偶然成功过的请求走缓存没真正发出去，能通；没缓存的新请求走全链路，死在第一步或第五步。API 表现完全不一致，排查时毫无规律可循。

### 5.5 最终方案

前端和后端各改一处。

**前端** `authFetch` 统一补尾斜杠，确保请求穿过 Astro 的 trailingSlash 检查：

```typescript
const finalUrl = url.includes('?') ? url : (url.endsWith('/') ? url : url + '/');
```

**后端** `main.go` 两处修改：

```go
// 关闭 Gin 的 301 重定向
r.RedirectTrailingSlash = false

// NoRoute 兜底：路由匹配失败时去掉尾斜杠再试一次
r.NoRoute(func(c *gin.Context) {
    path := c.Request.URL.Path
    if len(path) > 1 && strings.HasSuffix(path, "/") {
        c.Request.URL.Path = path[:len(path)-1]
        r.HandleContext(c)  // 重新走路由匹配
        return
    }
    c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
})
```

**为什么用 NoRoute 而不是 r.Use 中间件？** Gin 的执行顺序是：先匹配路由，再执行中间件链。如果在 Use 中间件里改 `c.Request.URL.Path`，路由已经匹配完了，404 已经返回了，来不及。`NoRoute` 在所有路由匹配失败后才执行，此时去掉斜杠再 `HandleContext` 重新走一遍，就能命中正确的 handler。

**最终请求链路**：

```
浏览器 fetch('/api/admin/stats')
  → authFetch 补 '/' → /api/admin/stats/
  → Astro 检查有斜杠，不予拦截
  → Vite proxy 匹配 '/api'，转发到 :8080
  → Gin 路由匹配失败（/stats/ 没注册）
  → NoRoute 触发 → 去斜杠 → /api/admin/stats
  → HandleContext 重新匹配 → 命中路由 → 200
```

### 5.6 教训

**当请求链路经过三个以上框架时，单独看每个框架的默认行为都没毛病，但它们叠加在一起会产生完全意想不到的效果。** 排查时一定要画链路图，搞清楚每一步请求经过了哪个组件、这个组件做了什么处理。

另外，**浏览器缓存 301 会让调试极其痛苦**。以后遇到类似情况，第一时间开无痕窗口验证，排除缓存干扰。

---

## 六、其他值得记录的坑

### 6.1 CSS 变量未定义导致按钮透明

后台的"添加管理员"按钮和对话框确认按钮用了 `bg-[var(--primary)]`。`--primary` 是前台主题配置文件里动态注入的 CSS 变量，后台页面没有加载这套注入逻辑。变量未定义，背景色回退为 `transparent`。

**解决**：后台按钮不用 CSS 变量，直接硬编码颜色值。实心按钮用 `bg-blue-600 hover:bg-blue-700 text-white`，边框按钮用 `border-gray-300 text-gray-700`。

### 6.2 GORM record not found 日志噪音

`AddAdmin` 里用 `First()` 查用户名是否已被占用。当用户名不存在时（这正是期望的结果——可以继续创建），GORM 会打一行 `record not found` 日志。这不是错误，最终 HTTP 返回 200，但看着闹心。

**解决**：用 `Count` 做查重不触发 not found：

```go
var count int64
s.DB.Model(&models.User{}).Where("username = ?", req.Username).Count(&count)
if count > 0 {
    return errors.New("用户名已被占用")
}
```

### 6.3 server.proxy 的「对象写法」和「字符串写法」

Vite proxy 配置中，`'/api': 'http://localhost:8080'` 和 `'/api': { target: 'http://localhost:8080', changeOrigin: true }` 都可以用。对象写法多了 `changeOrigin`（修改请求头的 Host 为目标地址），如果你的后端不做 Host 校验，两种写法效果一样。

---

## 七、服务端部署相关

### 7.1 systemd 环境变量导致 pnpm 找不到

后台的"触发重构"是 Go 通过 `exec.Command("sh", "-c", "pnpm build")` 执行的。本地跑没问题，但部署到服务器后 systemd 启动的服务 PATH 环境变量不包含 pnpm 的安装路径，报 `exit status 127: pnpm: command not found`。

**解决**：把 pnpm 软链接到 `/usr/local/bin/pnpm`，这个路径在 systemd 默认的 PATH 里。

### 7.2 上传代码后不要忘了重编译

新增后端路由后，只是上传了 Go 源码到服务器，还需要在服务器上：

```bash
cd /www/my-blog-project/backend
go build -o my-blog-backend .
sudo systemctl restart myblog-backend
```

否则跑的还是旧二进制，新注册的 `/api/admin/users` 等路由全部 404。

---

## 八、改动文件总览

从开始改造到现在的全部文件变动，方便以后回顾：

**后端（Go）**：

| 文件 | 操作 | 主要内容 |
|------|------|---------|
| `models/user.go` | 修改 | 追加 AddAdminRequest / ChangePasswordRequest / UserInfo 三个 DTO |
| `service/user_service.go` | 新建 | UserService：ListAdmins / AddAdmin / DeleteAdmin / ChangePassword |
| `controller/user_ctrl.go` | 新建 | UserController：四个接口 + Controller 层权限校验 |
| `service/post_service.go` | 修改 | SavePostResource 图片先存 preview-cache；ProcessPostPublish 搬运图片 + 路径替换 |
| `main.go` | 修改 | 依赖注入 + 用户路由注册 + RedirectTrailingSlash=false + NoRoute 去斜杠 + `/images/albums` 静态服务 |

**前端（Astro / TypeScript）**：

| 文件 | 操作 | 主要内容 |
|------|------|---------|
| `components/admin/Sidebar.astro` | 修改 | 移动端侧滑 + 遮罩 + 关闭按钮 + 用户管理导航 |
| `components/admin/UsersView.astro` | 新建 | 用户列表 + 添加/修改密码对话框 |
| `pages/admin/index.astro` | 修改 | 导入 UsersView + 汉堡菜单按钮 |
| `pages/admin/new-post.astro` | 修改 | 路径替换前端化移除 + 页面重置逻辑 |
| `pages/admin/login.astro` | 修改 | API 路径改为相对路径 |
| `scripts/admin-core.ts` | 修改 | 移动端菜单逻辑 + trailingSlash 适配 + 用户管理 CRUD |
| `astro.config.mjs` | 修改 | Vite proxy 配置（/api, /preview-cache, /images/albums） |
| `tailwind.config.cjs` | 修改 | fontFamily 移除 Roboto，让 body 内联字体生效 |
| `layouts/Layout.astro` | 修改 | import main.css，加载自定义字体声明 |

共 12 个文件。
