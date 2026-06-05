# 关于我

你好，我是 鲨鲨Don't。

目前还在上大学，主要使用 **Go语言** 。喜欢折腾莫名其妙的东西。

> 不可思议

## 关于本站

本站基于 [Mizuki](https://github.com/matsuzaka-yuki/mizuki) 主题搭建，前端框架为 **Astro**。Mizuki 原本是一个纯前端的博客主题，没有后端和管理后台。

::github{repo="matsuzaka-yuki/Mizuki"}

我对前端进行了大量个性化修改，包括但不限于：页面布局调整、样式重写、新增管理后台页面（文章编辑器、日记管理、相册管理、站点配置、用户管理等），以及移动端适配优化。

在此基础上，我**从零开发了 Go 语言后端**，采用 MVC 分层架构（Gin + GORM + MySQL），为博客提供了完整的内容管理能力：

- **文章管理**：Markdown 编辑器 + Frontmatter 解析，MySQL 与文件系统双写
- **日记/碎碎念**：JSON 文件存储，自动同步前端 TypeScript 数据文件
- **相册管理**：文件系统级增删改查，支持封面设置
- **站点配置**：可视化编辑全部主题设置
- **用户权限**：JWT 认证 + owner/admin 两级角色，支持多人协作
- **安全加固**：路径穿越防护、登录限流、JWT 吊销、文件类型校验、安全响应头

最终形成了一套完整的前后端分离博客系统，前端 Astro 静态生成，后端 Go 提供 API 和管理能力。

## 技术栈

- **前端框架**：Astro + Tailwind CSS + TypeScript
- **后端框架**：Go + Gin + GORM
- **数据库**：MySQL
- **编辑器**：Vditor（Markdown）
- **认证**：JWT（golang-jwt）

## 项目仓库

::github{repo="Asuka-20011204/Mizuki-Go-Blog"}

## 联系方式

- GitHub：[Asuka-20011204](https://github.com/Asuka-20011204)
- Bilibili：[鲨鲨Don&#39;t](https://space.bilibili.com/479044184)
- Gitee：[Asuka20011204](https://gitee.com/Asuka20011204)

## 致谢

感谢 [matsuzaka-yuki](https://github.com/matsuzaka-yuki) 开发了 Mizuki 主题，让我能在这个基础上加入自己的想法。
