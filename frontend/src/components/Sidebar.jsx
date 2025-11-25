import { useNavigate, useLocation, Outlet } from "react-router-dom";
import { Tabs } from "antd";
import { AndroidOutlined, AppleOutlined } from "@ant-design/icons";
import { useContext } from "react";
import { ThemeContext } from "../App";

const Sidebar = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { theme } = useContext(ThemeContext);
  
  // 根据主题设置背景色
  const isDark = theme === "dark" || (theme === "system" && window.matchMedia("(prefers-color-scheme: dark)").matches);

  const tabItems = [
    {
      key: "/",
      label: <span className="scbold text-base">首页</span>,
    },
    {
      key: "/logs",
      label: <span className="scbold text-base">日志</span>,
    },
        {
      key: "/wechat",
      label: <span className="scbold text-base">微信</span>,
    },
    {
      key: "/plugins",

      label: <span className="scbold text-base">插件</span>,
    },
    {
      key: "/settings",
      label: <span className="scbold text-base">设置</span>,
    },
  ];

  const handleTabChange = (key) => {
    navigate(key);
  };

  return (
    <div 
      className="flex flex-col h-screen"
      style={{
        backgroundColor: isDark ? "#141414" : "#ffffff",
        color: isDark ? "#ffffff" : "#000000",
      }}
    >
      <Tabs
        centered
        activeKey={location.pathname}
        items={tabItems}
        onChange={handleTabChange}
      />
      <div className="flex-1 overflow-hidden p-6">
        <Outlet />
      </div>
    </div>
  );
};

export default Sidebar;
