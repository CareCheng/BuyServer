// Package repository 数据访问层
// 本包封装所有数据库操作，提供统一的数据访问接口。
//
// 设计原则：
//   - 每个表对应一个 Repository 方法集
//   - 方法命名遵循 Get/Create/Update/Delete 前缀
//   - 复杂查询使用 Search/Find 前缀
//   - 统计查询使用 Count/Stats 前缀
//
// 主要功能：
//   - 用户数据访问（users 表）
//   - 订单数据访问（orders 表）
//   - 商品数据访问（products 表）
//   - 优惠券数据访问（coupons 表）
//   - 余额数据访问（balances, balance_logs 表）
//   - 积分数据访问（points, point_logs 表）
//   - 购物车数据访问（cart_items 表）
//   - 收藏数据访问（favorites 表）
//   - 工单数据访问（support_tickets, ticket_messages 表）
//   - 公告数据访问（announcements 表）
//   - FAQ数据访问（faqs, faq_categories 表）
//
// 使用示例：
//
//	repo := repository.NewRepository(db)
//
//	// 查询用户
//	user, err := repo.GetUserByID(1)
//
//	// 创建订单
//	err := repo.CreateOrder(&model.Order{...})
//
//	// 分页查询
//	orders, total, err := repo.GetOrdersByUserID(userID, 1, 20)
//
// 查询参数结构体：
// 复杂查询使用参数结构体，便于扩展和维护
//
//	type OrderSearchParams struct {
//	    UserID     *uint
//	    Status     *int
//	    StartDate  *time.Time
//	    EndDate    *time.Time
//	    Keyword    string
//	}
//
// 事务处理：
// Repository 层不处理事务，事务管理由 Service 层负责
//
//	tx := db.Begin()
//	repo := repository.NewRepository(tx)
//	// 执行多个操作
//	if err != nil {
//	    tx.Rollback()
//	    return err
//	}
//	tx.Commit()
package repository

// ==================== 查询优化说明 ====================
//
// 1. 索引使用
//    - 查询条件尽量使用索引字段
//    - 避免在索引字段上使用函数
//    - 联合索引注意字段顺序
//
// 2. 分页查询
//    - 使用 LIMIT + OFFSET 实现分页
//    - 大数据量使用游标分页
//    - 返回总数使用 COUNT(*) 单独查询
//
// 3. 预加载
//    - 使用 Preload 加载关联数据
//    - 避免 N+1 查询问题
//    - 只加载需要的字段
//
// 4. 缓存策略
//    - 热点数据使用内存缓存
//    - 配置信息使用数据库缓存
//    - 缓存失效使用主动刷新
//
// ==================== 常用查询模式 ====================
//
// 1. 精确查询
//    db.Where("id = ?", id).First(&model)
//
// 2. 模糊查询
//    db.Where("name LIKE ?", "%"+keyword+"%").Find(&models)
//
// 3. 范围查询
//    db.Where("created_at BETWEEN ? AND ?", start, end).Find(&models)
//
// 4. 分页查询
//    db.Offset((page-1)*pageSize).Limit(pageSize).Find(&models)
//
// 5. 排序查询
//    db.Order("created_at DESC").Find(&models)
//
// 6. 预加载查询
//    db.Preload("User").Find(&orders)
//
// 7. 聚合查询
//    db.Model(&Order{}).Select("SUM(price)").Where("status = ?", 2).Scan(&total)
