import { Routes, Route } from 'react-router-dom'
import Layout from './components/Layout'
import Home from './pages/Home'
import Articles from './pages/Articles'
import ArticleDetail from './pages/ArticleDetail'
import Archives from './pages/Archives'
import Categories from './pages/Categories'
import TagsPage from './pages/Tags'
import Unsubscribe from './pages/Unsubscribe'
import Projects from './pages/Projects'
import About from './pages/About'
import Search from './pages/Search'
import Login from './pages/Login'
import Register from './pages/Register'
import Publish from './pages/Publish'
import UserProfile from './pages/UserProfile'
import EditProfile from './pages/EditProfile'
import UserList from './pages/UserList'
import ProtectedRoute from './components/ProtectedRoute'

// Placeholder for pages not yet built
function ComingSoon({ title }: { title: string }) {
  return (
    <div style={{ textAlign: 'center', padding: 80 }}>
      <h2>{title}</h2>
      <p style={{ color: '#999' }}>This page is coming soon.</p>
    </div>
  )
}

export default function App() {
  return (
    <Routes>
      {/* ============ PUBLIC ROUTES ============ */}
      <Route element={<Layout />}>
        <Route path="/" element={<Home />} />
        <Route path="/articles" element={<Articles />} />
        <Route path="/article/:slug" element={<ArticleDetail />} />
        <Route path="/archives" element={<Archives />} />
        <Route path="/categories" element={<Categories />} />
        <Route path="/tags" element={<TagsPage />} />
        <Route path="/unsubscribe" element={<Unsubscribe />} />
        <Route path="/projects" element={<Projects />} />
        <Route path="/about" element={<About />} />
        <Route path="/search" element={<Search />} />
        <Route path="/user/:id" element={<UserProfile />} />
        <Route path="/user/:id/followers" element={<UserList type="followers" />} />
        <Route path="/user/:id/following" element={<UserList type="following" />} />

        {/* ============ PROTECTED ROUTES ============ */}
        <Route
          path="/publish"
          element={
            <ProtectedRoute action="publish articles">
              <Publish />
            </ProtectedRoute>
          }
        />
        <Route
          path="/article/:slug/edit"
          element={
            <ProtectedRoute action="edit articles">
              <Publish />
            </ProtectedRoute>
          }
        />
        <Route
          path="/settings"
          element={
            <ProtectedRoute action="access settings">
              <ComingSoon title="Settings" />
            </ProtectedRoute>
          }
        />
        <Route
          path="/profile"
          element={
            <ProtectedRoute action="view profile">
              <EditProfile />
            </ProtectedRoute>
          }
        />
        <Route
          path="/my-articles"
          element={
            <ProtectedRoute action="view your articles">
              <ComingSoon title="My Articles" />
            </ProtectedRoute>
          }
        />
      </Route>

      {/* ============ AUTH ROUTES — standalone ============ */}
      <Route path="/login" element={<Login />} />
      <Route path="/register" element={<Register />} />
    </Routes>
  )
}
