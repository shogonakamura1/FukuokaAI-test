# 依存関係図

## クリーンアーキテクチャの依存関係

```
┌─────────────────────────────────────────────────────────────────┐
│                         main.go                                  │
│  (依存関係の注入とルーティング設定)                                │
└────────────┬────────────────────────────────────────────────────┘
             │
             ├─────────────────────────────────────────────────────┐
             │                                                      │
             ▼                                                      ▼
┌────────────────────────────┐          ┌──────────────────────────────────────┐
│   Delivery Layer           │          │   Infrastructure Layer                │
│   (配信層)                  │          │   (インフラストラクチャ層)            │
├────────────────────────────┤          ├──────────────────────────────────────┤
│                            │          │                                      │
│  ┌──────────────────────┐ │          │  ┌──────────────────────────────┐   │
│  │ TripController       │ │          │  │ TripRepository (実装)        │   │
│  │                      │ │          │  │                              │   │
│  │ - CreateTrip()      │ │          │  │ implements                   │   │
│  │ - RecomputeTrip()   │ │          │  │ domain/repository            │   │
│  │ - GetShare()        │ │          │  │ .TripRepository             │   │
│  └──────────┬───────────┘ │          │  └──────────┬───────────────────┘   │
│             │              │          │             │                        │
│             │ depends on   │          │             │ implements             │
│             │              │          │             │                        │
└─────────────┼──────────────┘          └─────────────┼────────────────────────┘
              │                                       │
              │                                       │
              ▼                                       ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Usecase Layer                                 │
│                    (ユースケース層)                               │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ TripUsecase                                              │  │
│  │                                                          │  │
│  │ - CreateTrip()                                           │  │
│  │ - RecomputeTrip()                                        │  │
│  │ - GetShare()                                             │  │
│  │                                                          │  │
│  │ depends on:                                              │  │
│  │   - repository.TripRepository (interface)               │  │
│  │   - service.MLService (interface)                        │  │
│  └──────────────┬───────────────────────────────────────────┘  │
│                 │                                                │
│                 │ depends on                                     │
│                 │                                                │
└─────────────────┼────────────────────────────────────────────────┘
                  │
                  │
                  ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Domain Layer                                  │
│                    (ドメイン層)                                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │   Model      │  │  Repository  │  │   Service    │         │
│  │              │  │  Interface   │  │  Interface   │         │
│  │ - User       │  │              │  │              │         │
│  │ - Trip       │  │ TripRepository│ │ MLService    │         │
│  │ - TripPlace  │  │              │  │              │         │
│  │ - Share      │  │ (interface)  │  │ (interface)  │         │
│  │ - Route      │  │              │  │              │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
│                                                                  │
│  ⚠️ この層は他の層に依存しない（最内層）                          │
└─────────────────────────────────────────────────────────────────┘
```

## 依存関係の詳細

### 1. main.go
```
main.go
  ├─→ delivery/controller (TripController)
  ├─→ infrastructure/database (InitDB)
  ├─→ infrastructure/repository (TripRepository実装)
  ├─→ infrastructure/mlservice (MLService実装)
  └─→ usecase (TripUsecase)
```

### 2. Delivery Layer (delivery/controller)
```
TripController
  ├─→ usecase.TripUsecase
  └─→ domain/model (レスポンス用)
```

### 3. Usecase Layer (usecase)
```
TripUsecase
  ├─→ domain/repository.TripRepository (interface)
  ├─→ domain/service.MLService (interface)
  └─→ domain/model
```

### 4. Infrastructure Layer (infrastructure)
```
TripRepository (実装)
  ├─→ domain/repository.TripRepository (interfaceを実装)
  └─→ domain/model

MLService (実装)
  └─→ domain/service.MLService (interfaceを実装)

Database
  └─→ (外部依存のみ: sqlite)
```

### 5. Domain Layer (domain)
```
Entity
  └─→ (依存なし)

Repository Interface
  └─→ domain/model

Service Interface
  └─→ (依存なし)
```

## 依存関係の方向

```
外側 → 内側
  ↓
Delivery → Usecase → Domain
  ↓         ↓
Infrastructure (Domainのインターフェースを実装)
```

## 重要なポイント

1. **Domain層は最内層**
   - 他の層に一切依存しない
   - ビジネスロジックの核心

2. **Usecase層はDomain層のみに依存**
   - インターフェースに依存（実装には依存しない）
   - ビジネスロジックを実装

3. **Infrastructure層はDomain層のインターフェースを実装**
   - 実装の詳細を隠蔽
   - テスト時にモックに置き換え可能

4. **Delivery層はUsecase層に依存**
   - HTTPリクエスト/レスポンスの変換のみ
   - ビジネスロジックは含まない

5. **main.goはすべての層を統合**
   - 依存関係の注入（DI）
   - ルーティング設定

