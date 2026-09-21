import OperationLogTable from '../../components/OperationLogTable'

/** 安全日志（/logs/security）：只显示失败的操作用来排查异常 */
export default function SecurityLogs() {
  return (
    <OperationLogTable
      title="安全日志"
      subtitle="执行失败的敏感操作记录，用于排查越权、参数错误等问题"
      fixedStatus={false}
    />
  )
}
