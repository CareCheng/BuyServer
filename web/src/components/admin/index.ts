// 管理后台组件导出
export * from './types'

// 主要页面
export { DashboardPage } from './Dashboard'
export { ProductsPage } from './products/index'
export { CategoriesPage } from './Categories'
export { CouponsPage } from './Coupons'
export { OrdersPage } from './Orders'

// 组合页面（合并多个相关功能）
export { SystemManagePage } from './SystemManage'
export { SupportManagePage } from './SupportManage'
export { UserManagePage } from './UserManage'
export { ConfigManagePage } from './ConfigManage'
export { ContentManagePage } from './ContentManage'

// 首页配置
export { Homepage as HomepagePage } from './Homepage'

// 子页面（被组合页面引用）
export { UsersPage } from './Users'
export { AnnouncementsPage } from './Announcements'
export { PaymentPage } from './Payment'
export { EmailPage } from './Email'
export { DatabasePage } from './Database'
export { BackupsPage } from './Backups'
export { LogsPage } from './Logs'
export { SettingsPage } from './Settings'
export { Support as SupportPage } from './support/index'
export { StatsPage } from './Stats'
export { MonitorPage } from './Monitor'
export { RolesPage } from './Roles'
export { BalancePage } from './Balance'
export { InvoicesPage } from './Invoices'
export { BotPage } from './Bot'
export { ReviewsPage } from './Reviews'
export { KnowledgePage } from './Knowledge'
export { TicketTemplatesPage } from './TicketTemplates'
export { UndoPage } from './Undo'
export { DeletionsPage } from './Deletions'
export { PointsPage } from './Points'
export { TasksPage } from './Tasks'
export { FAQPage } from './FAQ'
