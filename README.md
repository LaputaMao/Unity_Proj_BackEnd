## Unity_Proj_BackEnd

### 目录结构

    unity_data_backend/
    ├── cmd/                  // 程序的主入口
    │   └── server/
    │       └── main.go       // Web服务器的启动文件
    │
    ├── internal/             // 项目内部的业务逻辑，外部无法引用
    │   ├── handler/              // API层，也叫handler，负责处理HTTP请求和响应
    │   ├── service/          // 服务层，处理核心业务逻辑（比如文件关系构建）
    │   ├── store/       // 仓库层，负责和数据库（MySQL）交互
    │   ├── router/            // 设置所有API路由
    │   └── model/            // 模型层，定义数据结构（比如文件信息的struct）
    │
    ├── configs/              // 配置文件目录 (比如数据库连接信息)
    ├── uploads/              // 用于存放用户上传的文件 (这个目录要记得加到.gitignore里)
    
    ├── go.mod                // Go模块依赖文件 (已存在)
    └── go.sum                // 依赖包的校验和文件 (会自动生成)
