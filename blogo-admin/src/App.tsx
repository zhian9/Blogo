import { Routes, Route, Navigate } from 'react-router-dom'
import AdminLayout from './components/AdminLayout'
import ProtectedRoute from './components/ProtectedRoute'
import './styles/global.css'
import LoginPage from './pages/login/LoginPage'
import Dashboard from './pages/dashboard/Dashboard'
import ArticleList from './pages/articles/ArticleList'
import ArticleEdit from './pages/articles/ArticleEdit'
import UserList from './pages/users/UserList'
import RoleList from './pages/roles/RoleList'
import MenuList from './pages/menus/MenuList'
import CategoryManage from './pages/categories/CategoryManage'
import TagManage from './pages/tags/TagManage'
import CommentManage from './pages/comments/CommentManage'
import PageList from './pages/settings/PageList'
import PageEdit from './pages/settings/PageEdit'
import SettingsManage from './pages/settings/SettingsManage'
import OperationLogs from './pages/logs/OperationLogs'
import LoginLogs from './pages/logs/LoginLogs'
import SecurityLogs from './pages/logs/SecurityLogs'
import AuditLog from './pages/logs/AuditLog'
import MediaLibrary from './pages/media/MediaLibrary'
import Profile from './pages/profile/Profile'
import Forbidden from './pages/errors/Forbidden'
import ProjectList from './pages/projects/ProjectList'
import ProjectEdit from './pages/projects/ProjectEdit'

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route
        element={
          <ProtectedRoute>
            <AdminLayout />
          </ProtectedRoute>
        }
      >
        <Route path="/" element={<Dashboard />} />
        <Route path="/articles" element={<ArticleList />} />
        <Route path="/articles/new" element={<ArticleEdit />} />
        <Route path="/articles/:id" element={<ArticleEdit />} />
        <Route path="/users" element={<UserList />} />
        <Route path="/roles" element={<RoleList />} />
        <Route path="/menus" element={<MenuList />} />
        <Route path="/categories" element={<CategoryManage />} />
        <Route path="/tags" element={<TagManage />} />
        <Route path="/projects" element={<ProjectList />} />
        <Route path="/projects/new" element={<ProjectEdit />} />
        <Route path="/projects/:id" element={<ProjectEdit />} />
        <Route path="/comments" element={<CommentManage />} />
        <Route path="/pages" element={<PageList />} />
        <Route path="/pages/new" element={<PageEdit />} />
        <Route path="/pages/:id" element={<PageEdit />} />
        <Route path="/settings" element={<SettingsManage />} />
        <Route path="/logs" element={<OperationLogs />} />
        <Route path="/logs/security" element={<SecurityLogs />} />
        <Route path="/media" element={<MediaLibrary />} />
        <Route path="/logs/audit" element={<AuditLog />} />
        <Route path="/logs/login" element={<LoginLogs />} />
        <Route path="/profile" element={<Profile />} />
        <Route path="/403" element={<Forbidden />} />
      </Route>
      <Route path="*" element={<Navigate to="/" />} />
    </Routes>
  )
}
