/**
 * 商品管理模块类型定义
 */

/**
 * 规格参数项接口
 */
export interface SpecItem {
  key: string
  value: string
}

/**
 * 商品表单数据
 */
export interface ProductFormData {
  name: string
  description: string
  price: string
  stock: string
  duration: string
  duration_unit: string
  status: string
  category_id: string
}

/**
 * 商品详情表单数据
 */
export interface DetailFormData {
  detail: string
  specs: SpecItem[]
  features: string[]
  tags: string
}
