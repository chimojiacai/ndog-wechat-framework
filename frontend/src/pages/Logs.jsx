import { Table, Tag, Tooltip, Button, App } from "antd";
import { useContext } from "react";
import { LogContext } from "../App";

const Logs = () => {
  // 从全局 Context 获取日志数据
  const { logs } = useContext(LogContext);
  // 使用 App hook 获取 message 实例
  const { message } = App.useApp();

  console.log("Logs 组件渲染，当前日志数量:", logs?.length || 0);
  console.log("日志数据:", logs);

  const columns = [
    {
      title: "时间",
      dataIndex: "timeStamp",
      key: "timeStamp",
      width: 120,
      align: "center",
      ellipsis: {
        showTitle: false,
      },
      render: (text) => (
        <Tooltip placement="topLeft" title={text}>
          {text}
        </Tooltip>
      ),
    },
    {
      title: "响应",
      dataIndex: "response",
      key: "response",
      width: 120,
      align: "center",
      ellipsis: {
        showTitle: false,
      },
      render: (text) => (
        <Tooltip placement="topLeft" title={text}>
          {text}
        </Tooltip>
      ),
    },
    {
      title: "类型",
      dataIndex: "type",
      key: "type",
      width: 60,
      align: "center",
      render: (text, record) => (
        <Tag color={record.color} style={{ margin: 0 }}>
          {text}
        </Tag>
      ),
    },

    {
      title: "消息",
      dataIndex: "msg",
      key: "msg",
      width: 200,
      align: "center",
      ellipsis: {
        showTitle: false,
      },
      render: (text) => (
        <Tooltip placement="topLeft" title={text}>
          {text}
        </Tooltip>
      ),
    },
    {
      title: "操作",
      key: "action",
      width: 60,
      align: "center",
      fixed: "right",
      render: (_, record) => (
        <Button
          variant="solid"
          size="small"
          onClick={() => {
            // 复制消息内容到剪贴板
            navigator.clipboard.writeText(record.msg).then(
              () => {
                message.success("已复制到剪贴板");
              },
              (err) => {
                message.error("复制失败");
                console.error("复制错误:", err);
              }
            );
          }}
        >
          复制
        </Button>
      ),
    },
  ];
  return (
    <div className="h-full flex flex-col overflow-hidden">
      <div className="flex-1 overflow-hidden">
        <Table
          columns={columns}
          dataSource={logs}
          rowKey={(record) => `${record.timeStamp}-${Math.random()}`}
          scroll={{ y: "calc(100vh - 184px)" }}
          pagination={{
            pageSize: 50,
            showTotal: (total) => `共 ${total} 条日志`,
          }}
          tableLayout="fixed"
          className="h-full"
        />
      </div>
    </div>
  );
};

export default Logs;
