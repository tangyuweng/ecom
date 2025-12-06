# E-Commerce Server API

一個使用 Go 語言開發的電商後端系統，採用 Clean Architecture 設計，提供完整的電商功能包括商品管理、購物車、訂單處理和用戶管理。

## 功能特色

- **用戶管理**

  - 用戶註冊與登入
  - JWT 令牌認證（Access Token + Refresh Token）
  - 個人資料管理
  - 密碼修改
  - 角色權限管理（用戶/管理員）

- **商品目錄**

  - 分類管理
  - 商品列表、搜尋、篩選、排序
  - 分頁支援
  - 庫存管理
  - 商品上下架功能

- **購物車系統**

  - 新增/修改/刪除購物車商品
  - 即時計算總金額
  - 購物車持久化

- **訂單管理**

  - 從購物車建立訂單
  - 訂單狀態追蹤（待處理 → 處理中 → 已出貨 → 已完成）
  - 訂單取消功能
  - 訂單歷史記錄
  - 自動庫存扣減

- **實時通知系統**

  - WebSocket 實時通知
  - 訂單狀態更新通知
  - 新訂單創建通知（管理員）
  - 訂單取消通知
  - 用戶連線管理

- **管理員功能**
  - 用戶管理
  - 商品 CRUD
  - 分類 CRUD
  - 訂單管理
  - 庫存管理

## 技術堆疊

- **語言**: Go 1.23
- **Web 框架**: Gin v1.11.0
- **ORM**: GORM v1.31.0
- **資料庫**: MySQL 8.0
- **認證**: JWT (golang-jwt/jwt v5.3.0)
- **WebSocket**: gorilla/websocket v1.5.3
- **文檔**: Swagger (swaggo)
- **容器化**: Docker & Docker Compose
- **遷移工具**: golang-migrate
- **密碼加密**: bcrypt

## 專案架構

本專案採用 Clean Architecture，確保程式碼的可維護性、可測試性和可擴展性：

```
server/
├── cmd/api/                    # 應用程式入口
│   └── main.go
├── conf/                       # 配置管理
│   └── config.go
├── internal/
│   ├── domain/                 # 領域層（核心業務邏輯）
│   │   ├── entities/          # 實體定義
│   │   └── repository/        # 資料庫介面定義
|   |   |__ service/           # 服務介面定義
│   ├── application/            # 應用層
│   │   ├── dtos/              # 資料傳輸物件
│   │   └── usecases/          # 用例（業務邏輯協調）
│   ├── infrastructure/         # 基礎設施層
│   │   ├── database/          # 資料庫實作
│   │   ├── jwt/               # JWT 服務
│   │   ├── migration/         # 資料庫遷移
│   │   └── seeding/           # 資料播種
|   |   |__ websocket/         # WebSocket服務
│   └── presentation/           # 表現層
│       ├── handlers/          # HTTP 處理器
│       ├── middleware/        # 中介軟體
│       └── router/            # 路由定義
├── docs/                       # Swagger 文檔
├── migrations/                 # SQL 遷移檔案
├── docker-compose.yml         # Docker 配置
├── Makefile                   # 建置自動化
└── .env                       # 環境變數配置
```

### 架構分層說明

- **Domain Layer（領域層）**: 包含核心業務實體（Entity）、值對象（Value Object）、資料庫介面（Repository）和領域服務介面（Service），不依賴任何框架
  - `entity/`: 領域實體與值對象（如 User, Product, Order, Notification）
  - `repository/`: 資料庫操作介面定義
  - `service/`: 領域服務介面（如 NotificationService, JWTService）
- **Application Layer（應用層）**: 協調業務邏輯，處理用例流程
  - `dto/`: 資料傳輸物件
  - `usecase/`: 業務用例實作
- **Infrastructure Layer（基礎設施層）**: 實作技術細節（資料庫、JWT、WebSocket、遷移等）
  - `database/`: Repository 實作
  - `jwt/`: JWT 服務實作
  - `websocket/`: WebSocket 服務實作
  - `migration/` & `seeding/`: 資料庫管理
- **Presentation Layer（表現層）**: 處理 HTTP 請求、路由和中介軟體
  - `handler/`: HTTP 處理器
  - `middleware/`: 認證、授權等中介軟體
  - `router/`: 路由配置

## 前置需求

- Go 1.23 或更高版本
- Docker 和 Docker Compose
- Make（用於執行 Makefile 指令）
- golang-migrate（用於資料庫遷移）

## 安裝與設定

### 1. 複製專案

```bash
git clone https://github.com/tangyuweng/ecom
cd server
```

### 2. 安裝相依套件

```bash
go mod download
```

### 3. 啟動 MySQL 資料庫

```bash
make docker-run
```

這將啟動一個 MySQL 8.0 容器：

- 容器端口：3306
- 主機端口：3307
- 資料庫名稱：ecom
- 使用者：ecom
- 密碼：ecom

### 4. 配置環境變數

複製 `.env.example` 為 `.env` 檔案：

```env
SERVER_PORT=:3000

# JWT 配置
JWT_SECRET=ecom-api-666
JWT_ACCESS_TOKE_EXP=24      # 小時
JWT_REFRESH_TOKE_EXP=72     # 小時

# MySQL 配置
MYSQL_HOST=localhost
MYSQL_PORT=3307
MYSQL_DBNAME=ecom
MYSQL_USERNAME=ecom
MYSQL_PASSWORD=ecom

# Cookie 配置
COOKIE_DOMAIN=localhost
COOKIE_SECURE=false          # HTTPS 環境設為 true
```

### 5. 執行資料庫遷移

```bash
make migrate-up
```

### 6. 播種測試資料（可選）

```bash
make seed
```

這將建立：

- 管理員帳號：`admin@example.com` / `admin123`
- 一般用戶：`user@example.com` / `user1234`
- 5 個商品分類
- 10 個範例商品

## 執行應用程式

### 開發模式

```bash
make run
```

或直接執行：

```bash
go run cmd/api/main.go
```

伺服器將在 `http://localhost:3000` 啟動

### 建置執行檔

```bash
make build
./bin/api
```

## API 文檔

啟動伺服器後，訪問 Swagger UI：

```
http://localhost:3000/swagger/index.html
```

### 重新生成 Swagger 文檔

```bash
make swagger
```

## 可用指令（Makefile）

| 指令                  | 說明                     |
| --------------------- | ------------------------ |
| `make swagger`        | 生成 Swagger API 文檔    |
| `make build`          | 編譯應用程式到 `bin/api` |
| `make run`            | 編譯並執行伺服器         |
| `make migrate-up`     | 執行所有資料庫遷移       |
| `make migrate-down`   | 回滾上一次遷移           |
| `make migrate-status` | 檢查遷移狀態             |
| `make seed`           | 播種測試資料             |
| `make docker-run`     | 啟動 MySQL 容器          |
| `make docker-stop`    | 停止 MySQL 容器          |

## API 端點總覽

### 認證 API (`/api/v1/auth`)

| 方法 | 端點             | 說明       | 需要認證 |
| ---- | ---------------- | ---------- | -------- |
| POST | `/auth/register` | 用戶註冊   | ✗        |
| POST | `/auth/login`    | 用戶登入   | ✗        |
| POST | `/auth/refresh`  | 刷新 Token | ✗        |
| POST | `/auth/logout`   | 登出       | ✗        |

### 用戶 API (`/api/v1/users`)

| 方法 | 端點                 | 說明         | 需要認證 |
| ---- | -------------------- | ------------ | -------- |
| GET  | `/users/me`          | 取得個人資料 | ✓        |
| PUT  | `/users/me`          | 更新個人資料 | ✓        |
| PUT  | `/users/me/password` | 修改密碼     | ✓        |

### 分類 API (`/api/v1/categories`)

| 方法 | 端點              | 說明         | 需要認證 |
| ---- | ----------------- | ------------ | -------- |
| GET  | `/categories`     | 取得所有分類 | ✗        |
| GET  | `/categories/:id` | 取得分類詳情 | ✗        |

### 商品 API (`/api/v1/products`)

| 方法 | 端點                     | 說明                                 | 需要認證 |
| ---- | ------------------------ | ------------------------------------ | -------- |
| GET  | `/products`              | 取得商品列表（支援篩選、排序、分頁） | ✗        |
| GET  | `/products/:id`          | 取得商品詳情                         | ✗        |
| GET  | `/products/category/:id` | 依分類取得商品                       | ✗        |

### 購物車 API (`/api/v1/cart`)

| 方法   | 端點              | 說明               | 需要認證 |
| ------ | ----------------- | ------------------ | -------- |
| GET    | `/cart`           | 取得購物車         | ✓        |
| POST   | `/cart/items`     | 加入商品到購物車   | ✓        |
| PUT    | `/cart/items/:id` | 更新購物車商品數量 | ✓        |
| DELETE | `/cart/items/:id` | 移除購物車商品     | ✓        |
| DELETE | `/cart`           | 清空購物車         | ✓        |

### 訂單 API (`/api/v1/orders`)

| 方法 | 端點                 | 說明             | 需要認證 |
| ---- | -------------------- | ---------------- | -------- |
| POST | `/orders`            | 從購物車建立訂單 | ✓        |
| GET  | `/orders`            | 取得訂單歷史     | ✓        |
| GET  | `/orders/:id`        | 取得訂單詳情     | ✓        |
| POST | `/orders/:id/cancel` | 取消訂單         | ✓        |

### 管理員 API (`/api/v1/admin`)

#### 用戶管理

| 方法  | 端點                    | 說明         | 需要管理員 |
| ----- | ----------------------- | ------------ | ---------- |
| GET   | `/admin/users`          | 取得所有用戶 | ✓          |
| PATCH | `/admin/users/:id/role` | 更新用戶角色 | ✓          |

#### 分類管理

| 方法   | 端點                    | 說明     | 需要管理員 |
| ------ | ----------------------- | -------- | ---------- |
| POST   | `/admin/categories`     | 建立分類 | ✓          |
| PUT    | `/admin/categories/:id` | 更新分類 | ✓          |
| DELETE | `/admin/categories/:id` | 刪除分類 | ✓          |

#### 商品管理

| 方法   | 端點                        | 說明     | 需要管理員 |
| ------ | --------------------------- | -------- | ---------- |
| POST   | `/admin/products`           | 建立商品 | ✓          |
| PUT    | `/admin/products/:id`       | 更新商品 | ✓          |
| DELETE | `/admin/products/:id`       | 刪除商品 | ✓          |
| PATCH  | `/admin/products/:id/stock` | 更新庫存 | ✓          |

#### 訂單管理

| 方法  | 端點                       | 說明         | 需要管理員 |
| ----- | -------------------------- | ------------ | ---------- |
| GET   | `/admin/orders`            | 獲得所有訂單 | ✓          |
| PATCH | `/admin/orders/:id/status` | 更新訂單狀態 | ✓          |

### WebSocket API (`/api/v1/ws`)

| 方法 | 端點               | 說明                 | 需要認證 |
| ---- | ------------------ | -------------------- | -------- |
| WS   | `/ws/notifications` | 訂單通知 WebSocket 連線 | ✓        |

## 認證機制

本系統使用 JWT（JSON Web Token）進行認證：

### Token 類型

1. **Access Token**

   - 有效期：24 小時（可配置）
   - 用於 API 請求認證
   - 放在 `Authorization` 標頭：`Bearer <access_token>`

2. **Refresh Token**
   - 有效期：72 小時（可配置）
   - 存儲在 HTTP-only Cookie 中
   - 用於刷新 Access Token

### 使用流程

1. **登入**：呼叫 `/auth/login` 取得 Access Token 和 Refresh Token
2. **API 請求**：在 `Authorization` 標頭帶上 `Bearer <access_token>`
3. **Token 過期**：使用 `/auth/refresh` 端點刷新 Access Token
4. **登出**：呼叫 `/auth/logout` 清除 Refresh Token

### 範例

```bash
# 登入
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"user1234"}'

# 使用 Access Token 呼叫 API
curl -X GET http://localhost:3000/api/v1/users/me \
  -H "Authorization: Bearer <your_access_token>"
```

## WebSocket 實時通知系統

本系統使用 WebSocket 提供實時訂單通知功能，讓用戶和管理員能即時收到訂單狀態變更。

### 通知領域模型

系統在 Domain Layer 定義了 `Notification` 值對象（`internal/domain/entity/notification.go`）：

- **Type**: 通知類型（`NotificationType`）
- **Payload**: 通知內容（彈性的 key-value 結構）

支援的通知類型：
- `order_status_updated`: 訂單狀態已更新
- `order_cancelled`: 訂單已取消
- `new_order_created`: 新訂單已建立（僅管理員）

### 通知觸發時機

1. **用戶建立訂單** → 所有管理員收到 `new_order_created` 通知
2. **管理員更新訂單狀態** → 訂單所屬用戶收到 `order_status_updated` 通知
3. **用戶取消訂單** → 用戶收到 `order_status_updated` 通知

### 架構設計

- **Domain Service**: `NotificationService` 介面定義通知行為（不依賴具體實作）
- **Domain Entity**: `Notification` 值對象封裝通知資料結構
- **Infrastructure**: `WsNotificationSvc` 實作 WebSocket 通知服務
- **Hub Pattern**: 管理所有 WebSocket 連線，支援單播和廣播

## 資料庫

### Schema

系統包含 7 個主要資料表：

- `users` - 用戶資料
- `categories` - 商品分類
- `products` - 商品資料
- `carts` - 購物車
- `cart_items` - 購物車項目
- `orders` - 訂單
- `order_items` - 訂單項目

### 遷移管理

```bash
# 執行所有遷移
make migrate-up

# 回滾上一次遷移
make migrate-down

# 檢查遷移狀態
make migrate-status
```

### 訂單狀態流程

```
Pending → Processing → Shipped → Completed
   ↓         ↓
    Cancelled
```

- 只有「待處理」和「處理中」狀態的訂單可以取消
- 管理員可以更新訂單狀態
- 建立訂單時會自動扣減商品庫存

## 開發指南

### 專案慣例

- 所有 ID 使用 UUID (CHAR(36))
- 密碼使用 bcrypt 加密
- 所有金額使用 DECIMAL(10,2)
- 使用 Repository Pattern 進行資料存取
- 遵循 Clean Architecture 原則

### 新增功能步驟

1. **Domain Layer**: 在 `internal/domain/entities/` 定義實體
2. **Repository Interface**: 在 `internal/domain/repository/` 定義資料庫介面
3. **Service Interface**: 在 `internal/domain/service/` 定義服務介面
4. **Infrastructure**: 在 `internal/infrastructure/database/` 實作 Repository
5. **DTOs**: 在 `internal/application/dtos/` 定義資料傳輸物件
6. **Use Cases**: 在 `internal/application/usecases/` 實作業務邏輯
7. **Handlers**: 在 `internal/presentation/handlers/` 實作 HTTP 處理器
8. **Router**: 在 `internal/presentation/router/` 註冊路由
9. **Swagger**: 添加 Swagger 註解並重新生成文檔

## 安全性

- 密碼使用 bcrypt 加密存儲
- JWT Token 用於身份驗證
- Refresh Token 存儲在 HTTP-only Cookie 中
- 角色基礎存取控制（RBAC）
- 輸入驗證和資料清理
- SQL 注入防護（使用 GORM 參數化查詢）

## 聯絡方式

如有問題或建議，請開 Issue 或 Pull Request。
