import OperationLogTable from '../../components/OperationLogTable'

/** 登录日志（/logs/login）：只显示认证模块的记录 */
export default function LoginLogs() {
  return (
    <OperationLogTable
      title="登录日志"
      subtitle="登录、注册与退出记录（含失败尝试）"
      fixedModule="认证模块"
      showModuleFilter={false}
    />
  )
}
