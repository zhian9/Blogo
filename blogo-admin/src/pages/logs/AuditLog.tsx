import OperationLogTable from '../../components/OperationLogTable'

/**
 * 安全审计 → 操作日志（菜单指向 /logs/audit）。
 * 记录全站关键操作：登录、注册、文章增删改、媒体上传等。
 */
export default function AuditLog() {
  return (
    <OperationLogTable
      title="操作日志"
      subtitle="全站关键操作审计记录（登录、内容变更、媒体操作等），支持按模块、操作人、结果与时间范围筛选"
    />
  )
}
