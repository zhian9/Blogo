import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react'
import type { ApiResponse, User, Role, Menu, Article, Category, Tag, Comment, Page, Setting, Logger, OperationLog, Statistics } from '../types'

const baseQuery = fetchBaseQuery({
  baseUrl: '/api/v1',
  prepareHeaders: (headers) => {
    const token = sessionStorage.getItem('admin-token')
    if (token) headers.set('Authorization', `Bearer ${token}`)
    return headers
  },
})

export const api = createApi({
  reducerPath: 'api',
  baseQuery,
  tagTypes: ['Articles', 'Users', 'Roles', 'Menus', 'Categories', 'Tags', 'Comments', 'Pages', 'Settings', 'Logs'],
  endpoints: (builder) => ({
    // ============ Articles ============
    getArticles: builder.query<ApiResponse<Article[]>, Record<string, any>>({
      query: (params) => ({ url: '/articles', params }),
      providesTags: ['Articles'],
    }),
    getArticle: builder.query<ApiResponse<Article>, string>({
      query: (id) => `/articles/${id}`,
      providesTags: (_r, _e, id) => [{ type: 'Articles', id }],
    }),
    createArticle: builder.mutation<ApiResponse<Article>, Partial<Article>>({
      query: (body) => ({ url: '/articles', method: 'POST', body }),
      invalidatesTags: ['Articles'],
    }),
    updateArticle: builder.mutation<ApiResponse<void>, { id: string; body: Partial<Article> }>({
      query: ({ id, body }) => ({ url: `/articles/${id}`, method: 'PUT', body }),
      invalidatesTags: ['Articles'],
    }),
    deleteArticle: builder.mutation<ApiResponse<void>, string>({
      query: (id) => ({ url: `/articles/${id}`, method: 'DELETE' }),
      invalidatesTags: ['Articles'],
    }),
    batchUpdateArticleStatus: builder.mutation<ApiResponse<void>, { ids: string[]; status: string }>({
      query: (body) => ({ url: '/articles/status', method: 'PATCH', body }),
      invalidatesTags: ['Articles'],
    }),
    toggleArticleTop: builder.mutation<ApiResponse<void>, { id: string; is_top: boolean }>({
      query: ({ id, is_top }) => ({ url: `/articles/${id}/top`, method: 'PATCH', body: { is_top } }),
      invalidatesTags: ['Articles'],
    }),

    // ============ Users ============
    getUsers: builder.query<ApiResponse<User[]>, Record<string, any>>({
      query: (params) => ({ url: '/users', params }),
      providesTags: ['Users'],
    }),
    createUser: builder.mutation<ApiResponse<User>, Partial<User>>({
      query: (body) => ({ url: '/users', method: 'POST', body }),
      invalidatesTags: ['Users'],
    }),
    updateUser: builder.mutation<ApiResponse<void>, { id: string; body: Partial<User> }>({
      query: ({ id, body }) => ({ url: `/users/${id}`, method: 'PUT', body }),
      invalidatesTags: ['Users'],
    }),
    deleteUser: builder.mutation<ApiResponse<void>, string>({
      query: (id) => ({ url: `/users/${id}`, method: 'DELETE' }),
      invalidatesTags: ['Users'],
    }),
    resetUserPassword: builder.mutation<ApiResponse<void>, string>({
      query: (id) => ({ url: `/users/${id}/reset-pwd`, method: 'PATCH' }),
    }),
    changeUserRole: builder.mutation<ApiResponse<void>, { id: string; role_code: string }>({
      query: ({ id, role_code }) => ({ url: `/users/role/${id}`, method: 'PUT', body: { role_code } }),
      invalidatesTags: ['Users'],
    }),
    changeUserStatus: builder.mutation<ApiResponse<void>, { id: string; status: string }>({
      query: ({ id, status }) => ({ url: `/users/status/${id}`, method: 'PUT', body: { status } }),
      invalidatesTags: ['Users'],
    }),

    // ============ Roles ============
    getRoles: builder.query<ApiResponse<Role[]>, Record<string, any>>({
      query: (params) => ({ url: '/roles', params }),
      providesTags: ['Roles'],
    }),
    createRole: builder.mutation<ApiResponse<Role>, Partial<Role>>({
      query: (body) => ({ url: '/roles', method: 'POST', body }),
      invalidatesTags: ['Roles'],
    }),
    updateRole: builder.mutation<ApiResponse<void>, { id: string; body: Partial<Role> }>({
      query: ({ id, body }) => ({ url: `/roles/${id}`, method: 'PUT', body }),
      invalidatesTags: ['Roles'],
    }),
    deleteRole: builder.mutation<ApiResponse<void>, string>({
      query: (id) => ({ url: `/roles/${id}`, method: 'DELETE' }),
      invalidatesTags: ['Roles'],
    }),

    // ============ Menus ============
    getMenus: builder.query<ApiResponse<Menu[]>, Record<string, any>>({
      query: (params) => ({ url: '/menus', params }),
      providesTags: ['Menus'],
    }),
    createMenu: builder.mutation<ApiResponse<Menu>, Partial<Menu>>({
      query: (body) => ({ url: '/menus', method: 'POST', body }),
      invalidatesTags: ['Menus'],
    }),
    updateMenu: builder.mutation<ApiResponse<void>, { id: string; body: Partial<Menu> }>({
      query: ({ id, body }) => ({ url: `/menus/${id}`, method: 'PUT', body }),
      invalidatesTags: ['Menus'],
    }),
    deleteMenu: builder.mutation<ApiResponse<void>, string>({
      query: (id) => ({ url: `/menus/${id}`, method: 'DELETE' }),
      invalidatesTags: ['Menus'],
    }),

    // ============ Categories ============
    getCategories: builder.query<ApiResponse<Category[]>, Record<string, any>>({
      query: (params) => ({ url: '/categories', params }),
      providesTags: ['Categories'],
    }),
    createCategory: builder.mutation<ApiResponse<Category>, Partial<Category>>({
      query: (body) => ({ url: '/categories', method: 'POST', body }),
      invalidatesTags: ['Categories'],
    }),
    updateCategory: builder.mutation<ApiResponse<void>, { id: string; body: Partial<Category> }>({
      query: ({ id, body }) => ({ url: `/categories/${id}`, method: 'PUT', body }),
      invalidatesTags: ['Categories'],
    }),
    deleteCategory: builder.mutation<ApiResponse<void>, string>({
      query: (id) => ({ url: `/categories/${id}`, method: 'DELETE' }),
      invalidatesTags: ['Categories'],
    }),

    // ============ Tags ============
    getTags: builder.query<ApiResponse<Tag[]>, Record<string, any>>({
      query: (params) => ({ url: '/tags', params }),
      providesTags: ['Tags'],
    }),
    createTag: builder.mutation<ApiResponse<Tag>, Partial<Tag>>({
      query: (body) => ({ url: '/tags', method: 'POST', body }),
      invalidatesTags: ['Tags'],
    }),
    updateTag: builder.mutation<ApiResponse<void>, { id: string; body: Partial<Tag> }>({
      query: ({ id, body }) => ({ url: `/tags/${id}`, method: 'PUT', body }),
      invalidatesTags: ['Tags'],
    }),
    deleteTag: builder.mutation<ApiResponse<void>, string>({
      query: (id) => ({ url: `/tags/${id}`, method: 'DELETE' }),
      invalidatesTags: ['Tags'],
    }),

    // ============ Comments ============
    getComments: builder.query<ApiResponse<Comment[]>, Record<string, any>>({
      query: (params) => ({ url: '/comments', params }),
      providesTags: ['Comments'],
    }),
    approveComment: builder.mutation<ApiResponse<void>, string>({
      query: (id) => ({ url: `/comments/${id}/approve`, method: 'PATCH' }),
      invalidatesTags: ['Comments'],
    }),
    rejectComment: builder.mutation<ApiResponse<void>, string>({
      query: (id) => ({ url: `/comments/${id}/reject`, method: 'PATCH' }),
      invalidatesTags: ['Comments'],
    }),
    deleteComment: builder.mutation<ApiResponse<void>, string>({
      query: (id) => ({ url: `/comments/${id}`, method: 'DELETE' }),
      invalidatesTags: ['Comments'],
    }),
    getCommentStats: builder.query<ApiResponse<{ total: number; pending: number; approved: number; rejected: number }>, void>({
      query: () => '/comments/stats',
      providesTags: ['Comments'],
    }),

    // ============ Pages ============
    getPages: builder.query<ApiResponse<Page[]>, Record<string, any>>({
      query: (params) => ({ url: '/pages', params }),
      providesTags: ['Pages'],
    }),
    createPage: builder.mutation<ApiResponse<Page>, Partial<Page>>({
      query: (body) => ({ url: '/pages', method: 'POST', body }),
      invalidatesTags: ['Pages'],
    }),
    updatePage: builder.mutation<ApiResponse<void>, { id: string; body: Partial<Page> }>({
      query: ({ id, body }) => ({ url: `/pages/${id}`, method: 'PUT', body }),
      invalidatesTags: ['Pages'],
    }),
    deletePage: builder.mutation<ApiResponse<void>, string>({
      query: (id) => ({ url: `/pages/${id}`, method: 'DELETE' }),
      invalidatesTags: ['Pages'],
    }),

    // ============ Settings ============
    getSettings: builder.query<ApiResponse<Setting[]>, void>({
      query: () => '/settings',
      providesTags: ['Settings'],
    }),
    updateSetting: builder.mutation<ApiResponse<void>, { key: string; body: Partial<Setting> }>({
      query: ({ key, body }) => ({ url: `/settings/${key}`, method: 'PUT', body }),
      invalidatesTags: ['Settings'],
    }),

    // ============ Logs ============
    getLogs: builder.query<ApiResponse<Logger[]>, Record<string, any>>({
      query: (params) => ({ url: '/loggers', params }),
      providesTags: ['Logs'],
    }),

    // ============ Operation Logs ============
    getOperationLogs: builder.query<ApiResponse<OperationLog[]>, Record<string, any>>({
      query: (params) => ({ url: '/operation-logs', params }),
      providesTags: ['Logs'],
    }),

    // ============ Statistics ============
    getStatistics: builder.query<ApiResponse<Statistics[]>, Record<string, any>>({
      query: (params) => ({ url: '/statistics', params }),
    }),

    // ============ Dashboard ============
    getTraffic: builder.query<ApiResponse<Statistics[]>, { days: number }>({
      query: ({ days }) => ({ url: '/dashboard/traffic', params: { days } }),
    }),
  }),
})

export const {
  useGetArticlesQuery, useGetArticleQuery, useCreateArticleMutation, useUpdateArticleMutation,
  useDeleteArticleMutation, useBatchUpdateArticleStatusMutation, useToggleArticleTopMutation,
  useGetUsersQuery, useCreateUserMutation, useUpdateUserMutation, useDeleteUserMutation,
  useResetUserPasswordMutation, useChangeUserRoleMutation, useChangeUserStatusMutation,
  useGetRolesQuery, useCreateRoleMutation, useUpdateRoleMutation, useDeleteRoleMutation,
  useGetMenusQuery, useCreateMenuMutation, useUpdateMenuMutation, useDeleteMenuMutation,
  useGetCategoriesQuery, useCreateCategoryMutation, useUpdateCategoryMutation, useDeleteCategoryMutation,
  useGetTagsQuery, useCreateTagMutation, useUpdateTagMutation, useDeleteTagMutation,
  useGetCommentsQuery, useApproveCommentMutation, useRejectCommentMutation, useDeleteCommentMutation, useGetCommentStatsQuery,
  useGetPagesQuery, useCreatePageMutation, useUpdatePageMutation, useDeletePageMutation,
  useGetSettingsQuery, useUpdateSettingMutation,
  useGetLogsQuery,
  useGetOperationLogsQuery,
  useGetStatisticsQuery, useGetTrafficQuery,
} = api
