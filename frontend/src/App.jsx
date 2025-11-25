import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import { ConfigProvider, App as AntApp, theme as antdTheme } from "antd";
import Sidebar from "./components/Sidebar";
import Home from "./pages/Home";
import Settings from "./pages/Settings";
import Logs from "./pages/Logs";
import Plugins from "./pages/Plugins";
// import My from "./pages/My";
import WechatList from "./pages/Wechat";
import { useMessageInit } from "./hooks/useNotification";

// 内部组件用于初始化 message
const AppContent = () => {
  useMessageInit();
  return (
    <Router
      future={{
        v7_startTransition: true,
        v7_relativeSplatPath: true,
      }}
    >
      <Routes>
        <Route path="/" element={<Sidebar />}>
          <Route index element={<Home />} />
          <Route path="settings" element={<Settings />} />
          <Route path="logs" element={<Logs />} />
          <Route path="plugins" element={<Plugins />} />
          <Route path="wechat" element={<WechatList />} />
        </Route>
      </Routes>
    </Router>
  );
};

import React, { useState, useEffect, createContext } from "react";
import { Events } from "@wailsio/runtime";
import { GetClearLog } from "../bindings/github.com/naidog/wechat-framework/service/config/configgetservice";

// 创建主题上下文
export const ThemeContext = createContext({
  theme: "light",
  setTheme: () => {},
});

// 创建日志上下文
export const LogContext = createContext({ logs: [], addLog: () => {} });

function App() {
  // 读取本地主题或默认
  const [theme, setTheme] = useState(
    () => localStorage.getItem("theme") || "light"
  );

  // 添加一个强制刷新的状态
  const [refreshKey, setRefreshKey] = useState(0);

  // 全局日志状态
  const [logs, setLogs] = useState([]);
  // 日志清理阈值
  const [clearLogThreshold, setClearLogThreshold] = useState(500);

  // 获取 clearLog 配置
  useEffect(() => {
    GetClearLog()
      .then((threshold) => {
        console.log("获取到 clearLog 配置:", threshold);
        setClearLogThreshold(threshold);
      })
      .catch((err) => {
        console.error("获取 clearLog 配置失败:", err);
      });
  }, []);

  // 全局监听日志事件（应用启动时就开始监听）
  useEffect(() => {
    console.log("全局开始监听日志事件...");

    const unsubscribe = Events.On("system:log", (event) => {
      console.log("全局收到日志事件:", event);

      // 提取真正的日志数据（从 WailsEvent 中提取）
      let logData = event;

      // 如果是 WailsEvent 对象，提取 data 字段
      if (event && event.data) {
        logData = Array.isArray(event.data) ? event.data[0] : event.data;
      }

      console.log("提取的日志数据:", logData);

      // 添加新日志到列表开头（最新的在最上面）
      setLogs((prevLogs) => {
        // 如果日志数量超过阈值，清空旧数据，只保留新日志
        if (prevLogs.length >= clearLogThreshold) {
          console.log(`日志数量超过阈值 ${clearLogThreshold}，清空旧数据`);
          return [logData];
        }

        const newLogs = [logData, ...prevLogs];
        console.log("更新后的日志数量:", newLogs.length);
        return newLogs;
      });
    });

    return () => {
      console.log("取消全局日志监听");
      unsubscribe();
    };
  }, [clearLogThreshold]);

  useEffect(() => {
    // storage事件用于多标签页同步
    const handler = (e) => {
      if (e.key === "theme" && e.newValue) setTheme(e.newValue);
    };
    window.addEventListener("storage", handler);
    // 自定义事件用于本页同步
    const themeChangeHandler = (e) => setTheme(e.detail);
    window.addEventListener("theme-change", themeChangeHandler);

    // 监听系统主题变化
    const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
    const systemThemeChangeHandler = (e) => {
      // 只有在设置为"跟随系统"时才触发更新
      const currentTheme = localStorage.getItem("theme");
      if (currentTheme === "system") {
        // 强制重新渲染以应用新的系统主题
        setRefreshKey((prev) => prev + 1);
      }
    };
    mediaQuery.addEventListener("change", systemThemeChangeHandler);

    return () => {
      window.removeEventListener("storage", handler);
      window.removeEventListener("theme-change", themeChangeHandler);
      mediaQuery.removeEventListener("change", systemThemeChangeHandler);
    };
  }, []);

  // 根据主题设置 body 和 root 背景色
  useEffect(() => {
    const isDark =
      theme === "dark" ||
      (theme === "system" &&
        window.matchMedia("(prefers-color-scheme: dark)").matches);
    const bgColor = isDark ? "#141414" : "#ffffff";

    document.body.style.backgroundColor = bgColor;
    const root = document.getElementById("root");
    if (root) {
      root.style.backgroundColor = bgColor;
    }
  }, [theme, refreshKey]); // 添加 refreshKey 依赖

  // 实时计算当前应该使用的主题算法
  const getCurrentThemeAlgorithm = () => {
    if (theme === "light") {
      return antdTheme.defaultAlgorithm;
    } else if (theme === "dark") {
      return antdTheme.darkAlgorithm;
    } else {
      // system - 实时检测系统主题
      return window.matchMedia("(prefers-color-scheme: dark)").matches
        ? antdTheme.darkAlgorithm
        : antdTheme.defaultAlgorithm;
    }
  };

  return (
    <LogContext.Provider value={{ logs }}>
      <ThemeContext.Provider value={{ theme, setTheme }}>
        <ConfigProvider
          theme={{
            token: {
              fontFamily: "'975Maru SC', sans-serif",
              fontSize: 14,
              fontSizeLG: 16,
              fontSizeSM: 12,
              fontWeightStrong: 600,
              colorPrimary: "#3959CF",
              borderRadius: 6,
            },
            algorithm: getCurrentThemeAlgorithm(),
          }}
        >
          <AntApp>
            <AppContent />
          </AntApp>
        </ConfigProvider>
      </ThemeContext.Provider>
    </LogContext.Provider>
  );
}

export default App;
