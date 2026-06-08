// ==================== API Response ====================

export interface ApiResponse<T = unknown> {
  success: boolean
  data: T
  total?: number
  error?: ApiError
}

export interface ApiError {
  id: string
  code: number
  detail: string
  status: string
}

// ==================== Auth ====================

export interface LoginRequest {
  username: string
  password: string
  captcha_id: string
  captcha_code: string
}

export interface LoginToken {
  access_token: string
  token_type: string
  expires_at: number
}

export interface CaptchaInfo {
  captcha_id: string
}

// ==================== User ====================

export interface User {
  id: string
  username: string
  name: string
  phone: string
  email: string
  avatar: string
  bio: string
  remark: string
  status: 'activated' | 'freezed'
  last_login_at: string
  last_login_ip: string
  follower_count: number
  following_count: number
  created_at: string
  updated_at: string
  roles: UserRole[]
}

export interface UserRole {
  id: string
  user_id: string
  role_id: string
  role?: Role
}

export interface UserForm {
  username: string
  name: string
  password?: string
  phone: string
  email: string
  remark: string
  status: 'activated' | 'freezed'
  roles: { role_id: string }[]
}

// ==================== Role ====================

export interface Role {
  id: string
  code: string
  name: string
  description: string
  sequence: number
  status: 'enabled' | 'disabled'
  created_at: string
  updated_at: string
  menus?: RoleMenu[]
}

export interface RoleMenu {
  id: string
  role_id: string
  menu_id: string
}

export interface RoleForm {
  code: string
  name: string
  description: string
  sequence: number
  status: 'enabled' | 'disabled'
  menus?: { menu_id: string }[]
}

// ==================== Menu ====================

export interface Menu {
  id: string
  code: string
  name: string
  description: string
  sequence: number
  type: 'page' | 'button'
  path: string
  properties: string
  status: 'enabled' | 'disabled'
  parent_id: string
  parent_path: string
  children?: Menu[]
  resources?: MenuResource[]
  created_at: string
  updated_at: string
}

export interface MenuResource {
  id: string
  menu_id: string
  path: string
  method: string
}

export interface MenuForm {
  code: string
  name: string
  description: string
  sequence: number
  type: 'page' | 'button'
  path: string
  properties: string
  status: 'enabled' | 'disabled'
  parent_id: string
  resources?: { path: string; method: string }[]
}

// ==================== Article ====================

export interface Article {
  id: string
  title: string
  slug: string
  summary: string
  content: string
  html_content: string
  cover_image_id: string
  category_id: string
  category?: Category
  author_id: string
  tags?: Tag[]
  views: number
  is_top: boolean
  status: 'draft' | 'published'
  visibility: 'public' | 'private' | 'partial_visible'
  visible_users?: { id: string; article_id: string; user_id: string }[]
  published_at: string
  seo_title: string
  seo_keywords: string
  seo_desc: string
  created_at: string
  updated_at: string
}

export interface ArticleForm {
  title: string
  slug: string
  summary: string
  content: string
  cover_image_id: string
  category_id: string
  tag_ids: string[]
  is_top: boolean
  status: 'draft' | 'published'
  visibility: 'public' | 'private' | 'partial_visible'
  visible_user_ids?: string[]
  published_at?: string
  seo_title: string
  seo_keywords: string
  seo_desc: string
}

// ==================== Category ====================

export interface Category {
  id: string
  name: string
  sort: number
  created_at: string
  updated_at: string
}

// ==================== Tag ====================

export interface Tag {
  id: string
  name: string
  created_at: string
  updated_at: string
}

// ==================== Comment ====================

export interface Comment {
  id: string
  article_id: string
  project_id: string
  user_id: string
  username: string
  email: string
  content: string
  status: 'approved' | 'pending' | 'rejected'
  parent_id: string
  is_top: boolean
  created_at: string
  updated_at: string
}

// ==================== Page ====================

export interface Page {
  id: string
  title: string
  slug: string
  content: string
  is_published: boolean
  created_at: string
  updated_at: string
}

export interface PageForm {
  title: string
  slug: string
  content: string
  is_published: boolean
}

// ==================== Setting ====================

export interface Setting {
  key: string
  value: string
  description: string
  created_at: string
  updated_at: string
}

// ==================== Logger ====================

export interface Logger {
  id: string
  level: string
  message: string
  context: string
  created_at: string
}

export interface OperationLog {
  id: string
  operator_id: string
  operator: string
  operator_ip: string
  module: string
  action_type: string
  description: string
  resource_id: string
  resource_name: string
  request_path: string
  request_method: string
  user_agent: string
  status: boolean
  status_code: number
  error_msg: string
  created_at: string
  updated_at: string
}

// ==================== Statistics ====================

export interface Statistics {
  id: string
  date: string
  pv: number
  uv: number
  ip_count: number
}

// ==================== Project ====================

export interface Project {
  id: string
  title: string
  slug: string
  summary: string
  content: string
  html_content: string
  cover_image_id: string
  cover_image?: Image
  category_id: string
  category?: Category
  author_id: string
  author?: User
  tags?: Tag[]
  views: number
  like_count: number
  favorite_count: number
  comment_count: number
  is_top: boolean
  is_featured: boolean
  featured_order: number
  status: 'draft' | 'published'
  visibility: 'public' | 'private' | 'partial_visible'
  project_state: 'developing' | 'completed' | 'maintaining' | 'paused' | 'archived'
  github_url: string
  demo_url: string
  visible_users?: { id: string; project_id: string; user_id: string }[]
  timeline?: ProjectTimeline[]
  resources?: ProjectResource[]
  published_at: string
  seo_title: string
  seo_keywords: string
  seo_desc: string
  created_at: string
  updated_at: string
}

export interface ProjectForm {
  title: string
  slug: string
  summary: string
  content: string
  cover_image_id: string
  category_id: string
  tag_ids: string[]
  is_top: boolean
  is_featured: boolean
  featured_order: number
  status: 'draft' | 'published'
  visibility: 'public' | 'private' | 'partial_visible'
  project_state: 'developing' | 'completed' | 'maintaining' | 'paused' | 'archived'
  github_url: string
  demo_url: string
  visible_user_ids?: string[]
  published_at?: string
  seo_title: string
  seo_keywords: string
  seo_desc: string
}

export interface ProjectTimeline {
  id: string
  project_id: string
  title: string
  description: string
  type: 'launch' | 'version' | 'feature' | 'milestone' | 'breaking' | 'archived'
  version: string
  image_id: string
  link: string
  event_date: string
  sort_order: number
  created_at: string
  updated_at: string
}

export interface ProjectTimelineForm {
  title: string
  description: string
  type: 'launch' | 'version' | 'feature' | 'milestone' | 'breaking' | 'archived'
  version: string
  image_id: string
  link: string
  event_date: string
  sort_order: number
}

export interface ProjectResource {
  id: string
  project_id: string
  title: string
  url: string
  type: 'document' | 'video' | 'slide' | 'article' | 'design' | 'other'
  sort_order: number
  created_at: string
  updated_at: string
}

export interface ProjectResourceForm {
  title: string
  url: string
  type: 'document' | 'video' | 'slide' | 'article' | 'design' | 'other'
  sort_order: number
}

export interface Image {
  id: string
  url: string
  path: string
  name: string
  size: number
  type: string
  width: number
  height: number
  category: string
}
