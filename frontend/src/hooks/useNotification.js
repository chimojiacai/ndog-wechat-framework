import { App } from "antd";

// 全局 message 实例
let messageApi = null;

// 初始化 Hook - 在 App.jsx 中调用一次
export const useMessageInit = () => {
  const { message } = App.useApp();
  messageApi = message;
};

// 全局 message 方法
export const msg = {
  success: (content) => {
    if (messageApi) {
      messageApi.success(content);
    }
  },
  error: (content) => {
    if (messageApi) {
      messageApi.error(content);
    }
  },
  info: (content) => {
    if (messageApi) {
      messageApi.info(content);
    }
  },
  warning: (content) => {
    if (messageApi) {
      messageApi.warning(content);
    }
  },
  loading: (content) => {
    if (messageApi) {
      return messageApi.loading(content);
    }
  },
};

export default msg;
