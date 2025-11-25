import { Card, Table, Tag, Space, Avatar, Tooltip, Button } from "antd";
import { UserOutlined } from "@ant-design/icons";
import { RunWechat } from "../../bindings/github.com/naidog/wechat-framework/service/wechat/wechatservice";
import { GetAccounts } from "../../bindings/github.com/naidog/wechat-framework/service/wechat/wechataccountservice";
import { msg } from "../hooks/useNotification";
import { Events } from "@wailsio/runtime";
import { useEffect, useState } from "react";
const WechatList = () => {
  const [accounts, setAccounts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [loginLoading, setLoginLoading] = useState(false);

  useEffect(() => {
    // 主动获取一次数据
    const fetchAccounts = async () => {
      try {
        // console.log("主动获取微信账号列表...");
        const data = await GetAccounts();
        // console.log("获取到账号:", data);
        setAccounts(data || []);
        setLoading(false);
      } catch (error) {
        // console.error("获取账号失败:", error);
        msg.error("获取账号失败: " + error);

        setLoading(false);
      }
    };

    fetchAccounts();

    // 监听后端推送的账号更新事件
    const unsubscribe = Events.On("wechat:accounts:update", (event) => {
      console.log("收到微信账号更新事件:", event);
      console.log(
        "event.data 类型:",
        typeof event?.data,
        "是否为数组:",
        Array.isArray(event?.data)
      );
      console.log("event.data 内容:", event?.data);

      // Wails 事件对象包含 data 属性
      let accounts = [];
      if (event && typeof event === "object" && "data" in event) {
        // WailsEvent 对象
        const data = event.data;
        // 检查 data 是否是嵌套数组
        if (Array.isArray(data) && data.length > 0 && Array.isArray(data[0])) {
          // 嵌套数组，取第一个元素
          accounts = data[0];
          console.log("检测到嵌套数组，扁平化处理");
        } else if (Array.isArray(data)) {
          // 正常数组
          accounts = data;
        }
      } else if (Array.isArray(event)) {
        // 直接是数组
        accounts = event;
      }

      console.log("最终账号数据:", accounts);
      console.log("账号数量:", accounts.length);
      if (accounts.length > 0) {
        console.log("第一个账号:", accounts[0]);
      }

      setAccounts(accounts);
      setLoading(false);
    });

    // 组件卸载时取消监听
    return () => {
      if (unsubscribe) {
        unsubscribe();
      }
    };
  }, []);

  const columns = [
    {
      title: "头像",
      dataIndex: "avatarUrl",
      key: "avatarUrl",
      width: 100,
      align: "center",
      render: (url) =>
        url ? (
          <Avatar
            shape="square"
            size="large"
            src={url}
            icon={<UserOutlined />}
          />
        ) : (
          <Avatar shape="square" size="large" icon={<UserOutlined />} />
        ),
    },
    {
      title: "昵称",
      dataIndex: "nick",
      key: "nick",
      width: 100,
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
      title: "微信号",
      dataIndex: "wxNum",
      key: "wxNum",
      width: 100,
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
      title: "微信ID",
      dataIndex: "wxid",
      key: "wxid",
      width: 100,
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
      title: "端口",
      dataIndex: "port",
      key: "port",
      width: 100,
      align: "center",
    },
    {
      title: "进程",
      dataIndex: "pid",
      key: "pid",
      width: 100,
      align: "center",
    },
    {
      title: "授权到期时间",
      dataIndex: "expireTime",
      key: "expireTime",
      width: 100,
      align: "center",
      ellipsis: {
        showTitle: false,
      },
      render: (time) => (
        <Tooltip placement="topLeft" title={time || "-"}>
          {time || "-"}
        </Tooltip>
      ),
    },
    {
      title: "状态",
      dataIndex: "isExpire",
      key: "isExpire",
      width: 100,
      align: "center",
      fixed: "right",
      render: (isExpire) => (
        <Tag color={isExpire === 0 ? "success" : "error"}>
          {isExpire === 0 ? "正常" : "已过期"}
        </Tag>
      ),
    },
  ];

  const loginWechat = async () => {
    if (loginLoading) return; // 如果正在加载，直接返回

    setLoginLoading(true);
    try {
      const result = await RunWechat();
      if (result) {
        msg.success("微信启动成功！");
      }
    } catch (error) {
      msg.error("启动失败：" + error);
    } finally {
      // 3秒后才能再次点击
      setTimeout(() => {
        setLoginLoading(false);
      }, 5000);
    }
  };
  return (
    <div className="h-full flex flex-col overflow-hidden">
      <div className="flex-1 overflow-hidden">
        <Table
          title={() => (
            <>
              <Button
                variant="solid"
                size="small"
                className="w-full"
                onClick={loginWechat}
                loading={loginLoading}
                disabled={loginLoading}
              >
                {loginLoading ? "启动中..." : "登录微信"}
              </Button>
              {/* <Tag className="!mt-4">
                特别说明：框架内置多开插件，请先确保所有微信已关闭，需要开几个微信，就点击几次多开，“全部点击完成后再逐个登录微信”！
              </Tag> */}
            </>
          )}
          bordered
          columns={columns}
          dataSource={accounts}
          rowKey="wxid"
          loading={loading}
          scroll={{ y: "calc(100vh - 242px)" }}
          pagination={{
            pageSize: 20,
            showTotal: (total) => `共 ${total} 个账号`,
          }}
          tableLayout="fixed"
          className="h-full"
        />
      </div>
    </div>
  );
};

export default WechatList;
