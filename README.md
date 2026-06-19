# CodeSense - AI驱动的智能代码仓库分析工具

[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](LICENSE)
[![OpenAI](https://img.shields.io/badge/OpenAI-API-412991?style=flat-square&logo=openai)](https://openai.com/)

**CodeSense** 是一个智能代码仓库分析工具，利用AI模型自动分析代码库结构、理解项目架构，并生成全面的技术报告。

## ✨ 核心特性

- 🤖 **AI智能分析**: 使用OpenAI模型（GPT-4、GPT-3.5）理解代码结构
- 📁 **智能代码探索**: 自动导航代码仓库，智能选择文件读取
- 🏗️ **架构识别**: 识别功能模块、包结构和依赖关系
- 📊 **全面报告**: 生成详细的Markdown技术分析报告
- 🔍 **上下文感知**: 维护对话历史，智能压缩上下文
- 🛡️ **安全保障**: 限制文件访问范围，防止目录穿越
- 🌐 **多语言支持**: 支持中文（zh）和英文（en）输出

## 📋 系统要求

- Go 1.26 或更高版本
- OpenAI API密钥（或兼容的API端点）

## 🚀 安装指南

### 使用Go Install安装
```bash
go install github.com/yuhua2000/codesense@latest
```

### 从源码构建
```bash
# 克隆仓库
git clone https://github.com/yuhua2000/codesense.git
cd codesense

# 构建项目
go build -o codesense main.go

# 或全局安装
go install
```

## 🎯 快速开始

### 基础使用
```bash
export OPENAI_API_KEY="你的API密钥"
codesense -path /你的/代码/路径 -model gpt-4-turbo
```

### 命令行参数
```bash
使用方法: codesense [选项]

选项:
  -path string
        代码仓库路径 (默认 ".")
  -model string
        AI模型名称 (如 "gpt-4-turbo", "gpt-3.5-turbo")
  -key string
        OpenAI API密钥 (可选，可使用OPENAI_API_KEY环境变量)
  -url string
        API基础URL (可选，默认为OpenAI官方端点)
  -lang string
        输出语言: zh (中文) 或 en (英文) (默认 "zh")
  -output string
        输出报告文件路径 (可选，默认输出到终端)
```

### 使用示例

1. **使用GPT-4分析当前目录：**
```bash
codesense -model gpt-4-turbo -lang en
```

2. **分析指定项目并保存到文件：**
```bash
codesense -path ~/projects/awesome-project -model gpt-4-turbo -output analysis.md
```

3. **使用自定义API端点：**
```bash
codesense -path . -model gpt-4-turbo -url "https://api.openai.com/v1/" -key "sk-..."
```

## 📁 项目结构

```
codesense/
├── main.go              # 主入口点和CLI接口
├── agent.go             # 代码分析智能代理，包含AI交互循环
├── openai.go           # OpenAI API客户端实现
├── file.go             # 文件系统操作（列出、读取、搜索）
├── prompt.go           # 分析提示词和模板
├── utils.go            # 工具函数和解析器
├── go.mod              # Go模块定义
└── go.sum              # 依赖校验和
```

## 🧠 工作原理

1. **目录扫描**: 首先扫描目标目录结构（默认扫描2层深度）
2. **AI代理初始化**: 创建带有分析提示的AI代理
3. **智能探索**: AI代理根据项目结构决定读取哪些文件
4. **上下文管理**: 维护并压缩对话历史，避免超出token限制
5. **报告生成**: 将分析结果合成完整的Markdown报告

### 分析过程
AI代理在执行过程中可以执行以下命令：
- `list_dir` - 列出目录内容
- `read_file` - 读取文件内容（支持大文件分块读取）
- `grep` - 在文件中搜索模式
- `finish` - 完成分析并生成最终报告

## 📊 报告示例

工具生成的报告包含以下部分：

```markdown
# 项目概述
• 项目描述和目的
• 核心功能和业务价值
• 技术架构概览

# 技术分析
• 编程语言和技术栈
• 框架和依赖分析
• 关键方法和入口点

# 功能清单
• 功能模块
• 模块职责和关系
• 关键实现方法

# 实现细节
• 关键目录和包
• 核心代码入口点
• 重要实现细节
```

### 📋 实际示例报告

查看 [report.md](./examples/report.md) CodeSense工具分析[gin-vue-admin/server](https://github.com/flipped-aurora/gin-vue-admin/tree/main/server)生成的实际报告。

## 🔧 配置说明

### 环境变量
```bash
export OPENAI_API_KEY="你的API密钥"
export OPENAI_BASE_URL="https://api.openai.com/v1/"  # 可选
```

### 模型选择
支持的模型（通过OpenAI API）：
- `gpt-4-turbo-preview`
- `gpt-4`
- `gpt-3.5-turbo`
- 其他兼容模型

## 📝 开发指南

### 从源码构建
```bash
# 克隆并构建
git clone https://github.com/yuhua2000/codesense.git
cd codesense
go build

# 运行测试
go test ./...
```

### 添加新功能
1. Fork仓库
2. 创建功能分支
3. 进行修改
4. 提交Pull Request

## 🤝 贡献指南

欢迎贡献！请随时提交Pull Request。

1. Fork本仓库
2. 创建功能分支 (`git checkout -b feature/新功能`)
3. 提交更改 (`git commit -m '添加新功能'`)
4. 推送到分支 (`git push origin feature/新功能`)
5. 提交Pull Request

## 📄 许可证

本项目采用MIT许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 致谢

- [OpenAI Go SDK](https://github.com/openai/openai-go) 提供OpenAI API客户端
- Go社区提供的优秀工具和库

## ⚠️ 限制说明

- 需要OpenAI API访问权限（或兼容API）
- 大型代码库受token限制
- 分析深度受API成本和性能限制
- 单个文件读取有大小限制

## 📞 支持与反馈

如有问题或建议：
1. 查看 [Issues](https://github.com/yuhua2000/codesense/issues) 页面
2. 如果问题未被解决，请创建新issue

---

**CodeSense** - 用AI让代码分析更智能 🚀"
