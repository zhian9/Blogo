// ==================== API Response Envelope ====================

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

export interface PaginationParams {
  current: number
  pageSize: number
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
  cover_image?: Image
  category_id: string
  category?: Category
  author_id: string
  author?: AuthorInfo
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

export interface ArchiveItem {
  year: number
  month: number
  count: number
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
  visibility?: 'public' | 'private' | 'partial_visible'
  visible_user_ids?: string[]
  seo_title: string
  seo_keywords: string
  seo_desc: string
}

// ==================== Category ====================

export interface Category {
  id: string
  name: string
  sort: number
  article_count?: number
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
  user_id: string
  username: string
  email: string
  content: string
  status: 'approved' | 'pending' | 'rejected'
  parent_id: string
  is_top: boolean
  created_at: string
  updated_at: string
  parent?: Comment
  user?: User
}

export interface CommentForm {
  article_id: string
  content: string
  parent_id?: string
  user_id?: string
  username?: string
  email?: string
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

// ==================== Setting ====================

export interface Setting {
  key: string
  value: string
  description: string
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
  created_at: string
  updated_at: string
}

// ==================== Image ====================

export interface Image {
  id: string
  url: string
  category: string
  filename: string
  size: number
  created_at: string
}

// ==================== User (from RBAC) ====================

export interface AuthorInfo {
  id: string
  username?: string
  name?: string
  avatar?: string
  bio?: string
  follower_count?: number
  following_count?: number
}

export interface User {
  id: string
  username: string
  name: string
  email?: string
  avatar?: string
}

// ==================== Auth ====================

export interface LoginRequest {
  username: string
  password: string
  captcha_id: string
  captcha_code: string
}

export interface RegisterRequest {
  username: string
  password: string
  confirm_password: string
  phone: string
  email: string
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

export interface AuthUser {
  id: string
  username: string
  name: string
  phone: string
  email: string
  avatar: string
  bio: string
  remark: string
  status: string
  last_login_at: string
  last_login_ip: string
  follower_count: number
  following_count: number
  created_at: string
  updated_at: string
  roles?: { id: string; role_id: string; role?: { id: string; code: string; name: string } }[]
}

// Contribution types
export interface ContributionDay {
  date: string
  count: number
  publish_count: number
  edit_count: number
  login_count: number
}

export interface ContributionStats {
  total_contributions: number
  current_streak: number
  longest_streak: number
  active_level: string
}

// Profile dashboard aggregation
export interface ProfileDashboard {
  user: AuthUser
  articles: Article[]
  liked_articles: Article[]
  favorite_articles: Article[]
  comments: Comment[]
  followers: string[]
  following: string[]
  contributions: ContributionDay[]
  contribution_stats: ContributionStats
}

export interface UpdateCurrentUser {
  name: string
  phone: string
  email: string
  avatar: string
  bio: string
  remark: string
}

// ==================== FriendLink ====================

export interface FriendLink {
  id: string
  name: string
  url: string
  description: string
  sort: number
}
