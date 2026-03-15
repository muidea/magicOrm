# VMI 实体定义和使用说明

## 文档概述

**生成时间**: 2026-01-24  

**文档目的**: 本文档整合了VMI系统的完整实体定义、业务场景说明、实体关系图和使用指南，为系统开发、数据建模和业务实施提供全面参考。

## 1. 业务场景说明

### 1.1 命名空间管理
namespace 在平台里统一管理，在业务使用时预先开通，并通过接入域名进行区分。例如 `autotest.remote.vpc` 对应的 namespace 为：`autotest`

### 1.2 状态管理
1. 状态信息包括ID、状态值、状态名
2. 由平台内置创建，不允许进行修改

### 1.3 多租户架构
所有实体都支持多租户隔离，通过 `namespace` 字段实现数据隔离和权限控制。

## 2. 实体概览

### 2.1 核心实体
- **partner** (会员信息, 路径:/vmi) - 系统会员实体
- **status** (状态路径:/vmi) - 通用状态实体

### 2.2 积分模块
- **credit** (积分信息, 路径:/vmi) - 积分记录
- **creditReport** (积分报表, 路径:/vmi/credit) - 积分统计报表
- **creditReward** (积分消费记录, 路径:/vmi/credit) - 积分支付记录
- **rewardPolicy** (积分策略, 路径:/vmi/credit) - 积分策略配置

### 2.3 订单模块
- **order** (订单信息, 路径:/vmi) - 客户订单
- **goodsItem** (商品条目, 路径:/vmi/order) - 订单商品项

### 2.4 产品模块
- **product** (产品信息, 路径:/vmi) - 产品定义
- **productInfo** (产品SKU, 路径:/vmi/product) - 产品SKU信息

### 2.5 店铺模块
- **store** (店铺, 路径:/vmi) - 店铺信息
- **member** (店铺成员, 路径:/vmi/store) - 店铺成员
- **goods** (商品信息, 路径:/vmi/store) - 店铺商品
- **goodsInfo** (出入库商品, 路径:/vmi/store) - 出入库商品
- **stockin** (入库单, 路径:/vmi/store) - 商品入库单
- **stockout** (出库单, 路径:/vmi/store) - 商品出库单

### 2.6 仓库模块
- **warehouse** (仓库信息, 路径:/vmi) - 仓库定义
- **shelf** (货架信息, 路径:/vmi/warehouse) - 仓库货架

## 3. 实体详细定义

### 3.1 status (状态)
- **id**: int64 (主键) - 唯一标识，由系统自动生成
- **value**: int (状态值) - 状态数值编码
- **name**: string (状态名称) - 状态显示名称

**业务说明**: 状态信息由平台内置创建，不允许进行修改。所有状态值通过统一的status实体进行管理。

### 3.2 partner (会员信息)
- **id**: int64 (主键) - 唯一标识，由系统自动生成
- **code**: string (会员码) - 唯一，由系统根据会员码生成规则自动生成
- **name**: string (姓名) - 必选
- **telephone**: string (电话) - 可选
- **wechat**: string (微信) - 可选
- **description**: string (描述) - 可选
- **referer**: partner* (推荐人) - 可选，自引用关系
- **status**: status* (状态) - 必选，由平台进行管理，允许进行更新
- **creater**: int64 (创建者) - 由系统自动生成
- **createTime**: int64 (创建时间) - 由系统自动生成
- **modifyTime**: int64 (修改时间) - 由系统自动更新
- **namespace**: string (命名空间) - 由系统自动生成

**业务说明**: 单个namespace里，可以拥有多个会员。推荐人关系为自引用，指向同一partner实体的其他记录。

### 3.3 credit (积分信息)
- **id**: int64 (主键) - 唯一标识，由系统自动生成
- **sn**: string (积分编号) - 唯一，由系统根据积分编号生成规则自动生成
- **owner**: partner* (所属会员) - 必选
- **memo**: string (备注) - 可选
- **credit**: int64 (积分) - 必选
- **type**: int (类型) - 必选
- **level**: int (等级) - 必选
- **creater**: int64 (创建者) - 由系统自动生成
- **createTime**: int64 (创建时间) - 由系统自动生成
- **namespace**: string (命名空间) - 由系统自动生成

**业务说明**: 单个namespace里，每个会员可以拥有多个积分信息。

### 3.4 creditReport (积分报表)
- **id**: int64 (主键) - 唯一标识，由系统自动生成
- **sn**: string (报表编号) - 唯一，由系统根据报表编号生成规则自动生成
- **owner**: partner* (所属会员) - 必选
- **credit**: int64 (累计积分) - 必选
- **available**: int64 (可用积分) - 必选
- **creater**: int64 (创建者) - 由系统自动生成
- **createTime**: int64 (创建时间) - 由系统自动生成
- **modifyTime**: int64 (修改时间) - 由系统自动更新
- **namespace**: string (命名空间) - 由系统自动生成

**业务说明**: 积分报表每个会员只能有一条有效记录。

### 3.5 creditReward (积分消费记录)
- **id**: int64 (主键) - 唯一标识，由系统自动生成
- **sn**: string (积分消费编号) - 唯一，由系统根据消费编号生成规则自动生成
- **owner**: partner* (所属会员) - 必选
- **credit**: int64 (消费积分值) - 必选
- **memo**: string (备注) - 可选
- **creater**: int64 (创建者) - 由系统自动生成
- **createTime**: int64 (创建时间) - 由系统自动生成
- **namespace**: string (命名空间) - 由系统自动生成

### 3.6 rewardPolicy (积分策略)
- **id**: int64 (主键) - 唯一标识，由系统自动生成
- **name**: string (策略名称) - 必选
- **description**: string (描述) - 可选
- **policy**: string (策略内容) - 必选
- **status**: status* (状态) - 必选，由平台进行管理，允许进行更新
- **creater**: int64 (创建者) - 由系统自动生成
- **createTime**: int64 (创建时间) - 由系统自动生成
- **namespace**: string (命名空间) - 由系统自动生成

### 3.7 order (订单信息)
- **id**: int64 (主键) - 唯一标识，由系统自动生成
- **sn**: string (订单编号) - 唯一，由系统根据订单号生成规则自动生成
- **type**: int (订单类型) - 必选(订单/退货)
- **customer**: partner* (客户) - 必选
- **goods**: goodsItem[] (商品列表) - 必选
- **cost**: float64 (总金额) - 必选
- **memo**: string (备注) - 可选
- **store**: store* (所属店铺) - 必选
- **status**: status* (状态) - 必选，由平台进行管理，允许进行更新
- **creater**: int64 (创建者) - 由系统自动生成
- **createTime**: int64 (创建时间) - 由系统自动生成
- **modifyTime**: int64 (修改时间) - 由系统自动更新
- **namespace**: string (命名空间) - 由系统自动生成

**业务说明**: 单个namespace里，可以拥有多个订单。

### 3.8 goodsItem (商品条目)
- **id**: int64 (主键) - 唯一标识，由系统自动生成
- **sku**: string (SKU编码) - 必选
- **name**: string (商品名称) - 必选
- **price**: float64 (单价) - 必选
- **count**: int (数量) - 必选
- **namespace**: string (命名空间) - 由系统自动生成

### 3.9 product (产品信息)
- **id**: int64 (主键) - 唯一标识，由系统自动生成
- **name**: string (产品名称) - 必选
- **description**: string (描述) - 可选
- **image**: string[] (图片) - 可选
- **expire**: int (有效期) - 可选
- **status**: status* (状态) - 必选，由平台进行管理，允许进行更新
- **tags**: string[] (标签) - 可选
- **creater**: int64 (创建者) - 由系统自动生成
- **createTime**: int64 (创建时间) - 由系统自动生成
- **modifyTime**: int64 (修改时间) - 由系统自动更新
- **namespace**: string (命名空间) - 由系统自动生成

**业务说明**: 单个namespace里，可以拥有多个产品。

### 3.10 productInfo (产品SKU)
- **id**: int64 (主键) - 唯一标识，由系统自动生成
- **sku**: string (SKU编码) - 必选
- **description**: string (描述) - 可选
- **image**: string[] (图片) - 可选
- **product**: product* (所属产品) - 必选，指向产品(product*)类型
- **creater**: int64 (创建者) - 由系统自动生成
- **createTime**: int64 (创建时间) - 由系统自动生成
- **modifyTime**: int64 (修改时间) - 由系统自动更新
- **namespace**: string (命名空间) - 由系统自动生成

### 3.11 store (店铺)
- **id**: int64 (主键) - 唯一标识，由系统自动生成
- **code**: string (店铺编码) - 唯一，由系统根据编码生成规则自动生成
- **name**: string (店铺名称) - 必选
- **description**: string (描述) - 可选
- **shelf**: shelf[] (货架列表) - 可选，包含一个货架列表，同一个货架不允许被两个店铺所拥有
- **creater**: int64 (创建者) - 由系统自动生成
- **createTime**: int64 (创建时间) - 由系统自动生成
- **modifyTime**: int64 (修改时间) - 由系统自动更新
- **namespace**: string (命名空间) - 由系统自动生成

**业务说明**: 单个namespace里，可以拥有多个店铺。

### 3.12 goods (商品信息)
- **id**: int64 (主键) - 唯一标识，由系统自动生成
- **sku**: string (SKU编码) - 必选
- **name**: string (商品名称) - 必选
- **description**: string (描述) - 可选
- **parameter**: string (参数) - 可选
- **serviceInfo**: string (服务信息) - 可选
- **product**: productInfo* (对应产品SKU) - 必选，指向产品SKU(productInfo*)类型
- **count**: int (库存数量) - 必选
- **price**: float64 (价格) - 必选
- **shelf**: shelf[] (所在货架) - 必选，为货架(shelf[])数组类型
- **store**: store* (所属店铺) - 必选
- **status**: status* (状态) - 必选，由平台进行管理，允许进行更新
- **creater**: int64 (创建者) - 由系统自动生成
- **createTime**: int64 (创建时间) - 由系统自动生成
- **modifyTime**: int64 (修改时间) - 由系统自动更新
- **namespace**: string (命名空间) - 由系统自动生成

### 3.13 goodsInfo (商品SKU)
- **id**: int64 (主键) - 唯一标识，由系统自动生成
- **sku**: string (SKU编码) - 必选
- **product**: productInfo* (对应产品SKU) - 必选，指向产品SKU(productInfo*)类型
- **type**: int (类型) - 必选(入库/出库)，由系统在进行入库/出库时进行自动设置
- **shelf**: shelf[] (所在货架) - 必选，为货架(shelf[])数组类型
- **count**: int (数量) - 必选
- **price**: float64 (价格) - 必选
- **creater**: int64 (创建者) - 由系统自动生成
- **createTime**: int64 (创建时间) - 由系统自动生成
- **namespace**: string (命名空间) - 由系统自动生成

### 3.14 member (店铺成员)
- **id**: int64 (主键) - 唯一标识，由系统自动生成
- **title**: string (职位) - 必选
- **name**: string (名称) - 必选
- **store**: store* (所属店铺) - 必选
- **creater**: int64 (创建者) - 由系统自动生成
- **createTime**: int64 (创建时间) - 由系统自动生成
- **modifyTime**: int64 (修改时间) - 由系统自动更新
- **namespace**: string (命名空间) - 由系统自动生成

### 3.15 stockin (入库单)
- **id**: int64 (主键) - 唯一标识，由系统自动生成
- **sn**: string (入库单号) - 唯一，由系统根据入库单号生成规则自动生成
- **goodsInfo**: goodsInfo[] (商品信息) - 必选
- **description**: string (描述) - 可选
- **status**: status* (状态) - 必选，由平台进行管理，允许进行更新
- **store**: store* (所属店铺) - 必选
- **creater**: int64 (创建者) - 由系统自动生成
- **createTime**: int64 (创建时间) - 由系统自动生成
- **modifyTime**: int64 (修改时间) - 由系统自动更新
- **namespace**: string (命名空间) - 由系统自动生成

**业务说明**: 入库单和出库单的状态由系统根据业务流程自动更新。商品库存数量在入库时增加。

### 3.16 stockout (出库单)
- **id**: int64 (主键) - 唯一标识，由系统自动生成
- **sn**: string (出库单号) - 唯一，由系统根据出库单号生成规则自动生成
- **goodsInfo**: goodsInfo[] (商品信息) - 必选
- **description**: string (描述) - 可选
- **status**: status* (状态) - 必选，由平台进行管理，允许进行更新
- **store**: store* (所属店铺) - 必选
- **creater**: int64 (创建者) - 由系统自动生成
- **createTime**: int64 (创建时间) - 由系统自动生成
- **modifyTime**: int64 (修改时间) - 由系统自动更新
- **namespace**: string (命名空间) - 由系统自动生成

**业务说明**: 商品库存数量在出库时减少。

### 3.17 warehouse (仓库信息)
- **id**: int64 (主键) - 唯一标识，由系统自动生成
- **code**: string (仓库编码) - 唯一，由系统根据编码生成规则自动生成
- **name**: string (仓库名称) - 必选
- **description**: string (描述) - 可选
- **creater**: int64 (创建者) - 由系统自动生成
- **createTime**: int64 (创建时间) - 由系统自动生成
- **modifyTime**: int64 (修改时间) - 由系统自动更新
- **namespace**: string (命名空间) - 由系统自动生成

**业务说明**: 单个namespace里，可以拥有多个仓库，每个仓库里存放多个货架，每个货架只属于一个仓库。

### 3.18 shelf (货架信息)
- **id**: int64 (主键) - 唯一标识，由系统自动生成
- **code**: string (货架编码) - 唯一，由系统根据编码生成规则自动生成
- **description**: string (描述) - 可选
- **used**: int (当前使用量) - 由系统根据业务使用量进行更新
- **capacity**: int (额定容量) - 由新建时指定，允许更新
- **warehouse**: warehouse* (所属仓库) - 在新建时指定，不允许进行修改且必选
- **status**: status* (状态) - 必选由平台进行管理，允许进行更新，标识该货架是否被使用
- **creater**: int64 (创建者) - 由系统自动生成
- **createTime**: int64 (创建时间) - 由系统自动生成
- **modifyTime**: int64 (修改时间) - 由系统自动更新
- **namespace**: string (命名空间) - 由系统自动生成

## 4. 实体关系图

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              VMI 实体关系图                                  │
│                  （箭头 A ───► B 表示 A 引用 B，即 A 包含 B 的引用）           │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                               核心实体模块                                   │
└─────────────────────────────────────────────────────────────────────────────┘

    [status] ◄─────────────────────────────────────────────────────────────┐
        │                                                                  │
        ▼                                                                  │
    [partner] ───┐ (自引用：referer)                                       │
        │        │                                                         │
        │        └─────────────────────────────────────────────────────┐   │
        ▼                                                              │   │
    ┌─────────────────────────────────────────────────────────────────────┐
    │                              积分模块                                │
    └─────────────────────────────────────────────────────────────────────┘
        │
        ├───► [credit] ────────► [creditReport]
        │         │                   
        │         ▼                   
        │     [creditReward]          
        │                             
        └───► [rewardPolicy]

    ┌─────────────────────────────────────────────────────────────────────┐
    │                              订单模块                                │
    └─────────────────────────────────────────────────────────────────────┘
        │
        └───► [order] ───► [goodsItem]
             │        │
             │        └───► [store]
             │
             └───► [partner] (客户)

    ┌─────────────────────────────────────────────────────────────────────┐
    │                              产品模块                                │
    └─────────────────────────────────────────────────────────────────────┘
        │
        └───► [product] ───► [productInfo]
             │
             └───► [status]

    ┌─────────────────────────────────────────────────────────────────────┐
    │                              店铺模块                                │
    └─────────────────────────────────────────────────────────────────────┘
        │
        └───► [store] ───┬───► [goods] ───┬───► [productInfo]
             │           │                ├───► [shelf]
             │           │                └───► [status]
             │           │
             │           ├───► [member]
             │           │
             │           ├───► [shelf]
             │           │
             │           ├───► [stockin] ───┬───► [goodsInfo]
             │           │                  └───► [status]
             │           │
             │           └───► [stockout] ───┬───► [goodsInfo]
             │                               └───► [status]
             │
             └───► [goodsInfo] ───┬───► [productInfo]
                                  ├───► [shelf]
                                  └───► [status]

    ┌─────────────────────────────────────────────────────────────────────┐
    │                              仓库模块                                │
    └─────────────────────────────────────────────────────────────────────┘
        │
        └───► [warehouse] ◄─── [shelf] ───► [status]

┌─────────────────────────────────────────────────────────────────────────────┐
│                           关键跨模块关系说明                                 │
└─────────────────────────────────────────────────────────────────────────────┘

1. status 被多个实体引用：partner, product, goods, goodsInfo, order, stockin, stockout, shelf, rewardPolicy
2. partner 被引用：credit, creditReport, creditReward, order (作为客户)
3. product 被引用：productInfo
4. productInfo 被引用：goods, goodsInfo
5. store 被引用：goods, member, stockin, stockout, order, shelf, goodsInfo
6. shelf 被引用：goods, goodsInfo, store, warehouse
7. goodsInfo 被引用：stockin, stockout
8. 所有实体都包含 creater, createTime, namespace 通用字段，部分实体包含 modifyTime 字段
```

## 5. 关系说明

### 5.1 核心引用关系
1. **partner → status**: 会员状态
2. **partner → partner**: 推荐人关系（自引用）

### 5.2 积分模块关系
1. **credit → partner**: 积分所属会员
2. **creditReport → partner**: 报表所属会员
3. **creditReward → partner**: 积分记录所属会员
4. **rewardPolicy → status**: 策略状态

### 5.3 订单模块关系
1. **order → partner**: 订单客户
2. **order → store**: 订单所属店铺
3. **order → status**: 订单状态
4. **order → goodsItem**: 订单商品项

### 5.4 产品模块关系
1. **product → status**: 产品状态
2. **product → productInfo**: 产品SKU信息（一对多）

### 5.5 店铺模块关系
1. **store → goods**: 店铺商品（一对多）
2. **store → member**: 店铺成员（一对多）
3. **store → shelf**: 店铺货架（一对多）
4. **store → stockin**: 入库单（一对多）
5. **store → stockout**: 出库单（一对多）
6. **goods → productInfo**: 商品对应产品SKU
7. **goods → shelf**: 商品所在货架（一对多）
8. **goods → status**: 商品状态
9. **goodsInfo → productInfo**: 商品SKU对应产品SKU
10. **goodsInfo → shelf**: 商品SKU所在货架（一对多）
11. **goodsInfo → status**: 商品SKU状态
12. **goodsInfo → store**: 商品SKU所属店铺
13. **member → store**: 成员所属店铺
14. **stockin → goodsInfo**: 入库商品（一对多）
15. **stockin → status**: 入库单状态
16. **stockin → store**: 入库单所属店铺
17. **stockout → goodsInfo**: 出库商品（一对多）
18. **stockout → status**: 出库单状态
19. **stockout → store**: 出库单所属店铺

### 5.6 仓库模块关系
1. **shelf → warehouse**: 货架所属仓库
2. **shelf → status**: 货架状态

### 5.7 跨模块引用关系
1. **status 被广泛引用**: 被 partner, product, goods, goodsInfo, order, stockin, stockout, shelf, rewardPolicy 引用
2. **partner 被引用**: 被 credit, creditReport, creditReward, order 引用
3. **product 被跨模块引用**: 被 productInfo 引用
4. **productInfo 被跨模块引用**: 被 goods, goodsInfo 引用
5. **store 被跨模块引用**: 被 goods, member, stockin, stockout, order, shelf, goodsInfo 引用
6. **shelf 被跨模块引用**: 被 goods, goodsInfo, store, warehouse 引用

## 6. 模块划分

### 6.1 会员管理模块
- partner, status

### 6.2 积分管理模块
- credit, creditReport, creditReward, rewardPolicy

### 6.3 订单交易模块
- order, goodsItem

### 6.4 产品管理模块
- product, productInfo

### 6.5 店铺运营模块
- store, goods, goodsInfo, member, stockin, stockout

### 6.6 仓储管理模块
- warehouse, shelf

## 7. 系统特性

### 7.1 通用字段
所有实体都包含以下通用字段：
- creater: 创建者ID
- createTime: 创建时间戳
- namespace: 命名空间（多租户支持）

部分实体包含以下字段：
- modifyTime: 修改时间戳（仅在有更新需求的实体中包含）

### 7.2 状态管理
- 使用统一的status实体管理各种状态
- 支持状态值的灵活配置

### 7.3 编码生成
多个实体支持自动编码生成：
- partner.code: 会员码
- credit.sn: 积分编号
- creditReport.sn: 报表编号
- creditReward.sn: 积分记录编号
- store.code: 店铺编码
- stockin.sn: 入库单号
- stockout.sn: 出库单号
- warehouse.code: 仓库编码
- shelf.code: 货架编码

### 7.4 数据关系
- 支持一对多关系（数组类型）
- 支持多对一引用（指针类型）
- 支持自引用关系（partner.referer）

## 8. 业务规则

### 8.1 状态管理规则
- 状态信息由平台内置创建，不允许进行修改
- 所有状态值通过统一的status实体进行管理

### 8.2 编码生成规则
- 会员码、积分编号、报表编号等编码由系统根据编码生成规则自动生成
- 编码在命名空间内保持唯一性

### 8.3 字段约束规则
- 创建人、创建时间、命名空间由系统自动生成，不可修改
- 修改时间字段在数据更新时由系统自动更新
- 必选字段在创建时必须提供，可选字段可以为空

### 8.4 数据关系约束
- 同一个货架不允许被两个店铺所拥有
- 商品SKU的type字段标识入库/出库类型，由系统在进行入库/出库操作时自动设置
- 推荐人关系为自引用，指向同一partner实体的其他记录

### 8.5 业务逻辑规则
- 积分报表每个会员只能有一条有效记录
- 入库单和出库单的状态由系统根据业务流程自动更新
- 商品库存数量在入库时增加，出库时减少

## 9. 使用说明

### 9.1 系统初始化流程
1. **创建命名空间**: 为每个业务租户创建独立的namespace
2. **初始化状态数据**: 导入平台内置的状态配置
3. **创建仓库和货架**: 建立仓储基础设施
4. **创建店铺**: 设置运营店铺
5. **创建产品**: 定义产品线和SKU
6. **创建会员**: 注册系统用户

### 9.2 日常业务流程
1. **会员注册**: 创建partner记录，自动生成会员码
2. **积分管理**:
   - 会员注册时根据rewardPolicy分配初始积分
   - 订单完成后根据orderCredit规则计算并添加积分
   - 积分消费时创建creditReward记录
3. **订单处理**:
   - 创建order记录，关联客户、店铺和商品
   - 更新库存数量
   - 计算并更新会员积分
4. **库存管理**:
   - 入库操作创建stockin记录，增加库存
   - 出库操作创建stockout记录，减少库存
   - 实时更新goods.count字段

## 10. 最佳实践

### 10.1 命名空间设计
- 建议按业务线或客户群体划分namespace
- 每个namespace应有明确的业务边界和数据隔离需求
- 定期清理不再使用的namespace以释放资源

### 10.2 状态管理
- 使用status实体统一管理所有状态，避免硬编码状态值
- 为每种业务场景定义清晰的状态流转规则
- 定期审核状态配置，确保业务逻辑的一致性

### 10.3 编码规范
- 编码生成规则应保证全局唯一性
- 建议使用有意义的编码前缀（如：MEM-会员，ORD-订单）
- 编码长度应适中，便于人工识别和处理

### 10.4 性能优化
- 为频繁查询的字段建立索引（如：namespace, owner, store等）
- 定期清理历史数据，控制单表数据量
- 使用分库分表策略应对大数据量场景

### 10.5 数据一致性
- 重要业务操作使用事务保证数据一致性
- 建立数据校验机制，防止脏数据产生
- 实现数据变更审计，记录关键字段的修改历史

## 11. 总结

VMI系统包含6个主要模块，共计18个实体，形成了完整的电商管理系统实体模型。实体之间通过清晰的引用关系连接，支持会员管理、订单处理、产品管理、店铺运营、仓储管理等核心业务功能。

### 11.1 核心价值
1. **模块化设计**: 清晰的模块划分，便于系统扩展和维护
2. **多租户支持**: 通过namespace实现数据隔离，支持多客户部署
3. **灵活配置**: 状态、积分策略等均可配置，适应不同业务需求
4. **完整链路**: 覆盖从会员注册到订单完成的完整业务链路

### 11.2 适用场景
- 电商平台会员积分系统
- 零售店铺库存管理系统
- 多店铺连锁运营平台
- 产品SKU管理系统

### 11.3 后续扩展建议
1. **数据分析模块**: 添加销售报表、库存分析等数据统计功能
2. **权限管理**: 细化店铺成员的操作权限控制
3. **消息通知**: 集成订单状态变更、库存预警等消息通知
4. **API扩展**: 提供更丰富的API接口，支持第三方系统集成

---
**文档版本**: 1.0
**最后更新**: 2026-01-24
**维护团队**: VMI开发团队