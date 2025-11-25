import { useState, useEffect } from "react";
// import { Window } from "@wailsio/runtime";
import { GetWechatConfig } from "../../bindings/github.com/naidog/wechat-framework/service/config/configgetservice";
import { SetWechatConfig } from "../../bindings/github.com/naidog/wechat-framework/service/config/configsetservice";
import { GetWechatPaths } from "../../bindings/github.com/naidog/wechat-framework/service/utils/getwechatpathservice";
import { NoupdateWechat } from "../../bindings/github.com/naidog/wechat-framework/service/utils/noupdatewechatservice";
import {
  GetTheme,
  SetTheme,
} from "../../bindings/github.com/naidog/wechat-framework/service/config/themeservice";
import { msg } from "../hooks/useNotification";
import { QuestionCircleOutlined } from "@ant-design/icons";
import {
  Tooltip,
  Badge,
  Select,
  Input,
  Card,
  Space,
  Button,
  Checkbox,
  InputNumber,
  Radio,
} from "antd";
const Settings = () => {
  const [wechatConfig, setWechatConfig] = useState(null);
  const [wechatPath, setWechatPath] = useState(null);
  const [theme, setTheme] = useState("light");

  // 表单值更新函数
  const onChangeLogs = (e) => {
    setWechatConfig({ ...wechatConfig, logs: e.target.checked });
  };
  const handleChangeGroupMemberEvent = (value) => {
    setWechatConfig({ ...wechatConfig, groupMemberEvent: value });
  };
  const handleChangeHookSilkEvent = (value) => {
    setWechatConfig({ ...wechatConfig, hookSilk: value });
  };
  const handleChangeResend = (value) => {
    setWechatConfig({ ...wechatConfig, resend: value });
  };
  const onChangeTimeOut = (value) => {
    setWechatConfig({ ...wechatConfig, timeOut: Math.floor(value * 1000) });
  };
  const onChangeIgnoreMsg = (value) => {
    setWechatConfig({ ...wechatConfig, ignoreMsg: value });
  };
  const onChangeClearLog = (value) => {
    setWechatConfig({ ...wechatConfig, clearLog: value });
  };

  const onChangeInstallPath = (e) => {
    setWechatConfig({ ...wechatConfig, installationPath: e.target.value });
  };

  const onChangeCachePath = (e) => {
    setWechatConfig({ ...wechatConfig, cachePath: e.target.value });
  };

  const handleChangeVersion = (value) => {
    setWechatConfig({ ...wechatConfig, version: [value] });
  };

  const handleChangeUpdate = (value) => {
    setWechatConfig({ ...wechatConfig, update: value });
  };

  // 主题切换处理
  const handleThemeChange = async (e) => {
    const newTheme = e.target.value;
    try {
      const result = await SetTheme(newTheme);
      if (result) {
        setTheme(newTheme);
        localStorage.setItem("theme", newTheme);
        window.dispatchEvent(
          new CustomEvent("theme-change", { detail: newTheme })
        );

        msg.success("主题设置成功！");
      }
    } catch (error) {
      msg.error("主题设置失败：" + error);
    }
  };

  const getWechatPath = async () => {
    try {
      const path = await GetWechatPaths();
      setWechatPath(path);
      // 同时更新配置中的安装目录和缓存目录
      setWechatConfig({
        ...wechatConfig,
        installationPath: path.installationPath || "",
        cachePath: path.cachePath || "",
      });

      console.log("获取目录成功", path);
    } catch (error) {
      msg.error("获取微信目录失败：" + error);
    }
  };

  // 保存配置
  const handleSaveConfig = async () => {
    try {
      const result = await SetWechatConfig(wechatConfig);
      if (result) {
        msg.success("保存成功！为确保部分设置立即生效，建议重启框架和微信。");
        NoupdateWechat();
      }
    } catch (error) {
      msg.error("保存失败：" + error);
    }
  };

  // 刷新配置
  const handleRefreshConfig = async () => {
    try {
      const config = await GetWechatConfig();
      setWechatConfig(config);
      msg.success("刷新成功！");
      console.log(config);
    } catch (error) {
      msg.error("刷新失败：" + error);
    }
  };
  useEffect(() => {
    const getWechatConfig = async () => {
      try {
        const config = await GetWechatConfig();
        setWechatConfig(config);
        // console.log("微信配置:", config);
      } catch (error) {
        msg.error("获取配置文件失败：" + error);
      }
    };

    const getTheme = async () => {
      try {
        const currentTheme = await GetTheme();
        setTheme(currentTheme || "light");
      } catch (error) {
        console.error("获取主题失败：", error);
      }
    };

    getWechatConfig();
    getTheme();
    NoupdateWechat();
  }, []);
  return (
    <div
      className="h-full overflow-auto flex flex-col gap-4"
      style={{
        scrollbarWidth: "none", // Firefox
        msOverflowStyle: "none", // IE/Edge
      }}
    >
      <Card size="small">
        <Space>
          <Button variant="solid" size="small" onClick={handleSaveConfig}>
            保存修改
          </Button>

          <Button variant="solid" size="small" onClick={handleRefreshConfig}>
            刷新设置
          </Button>
          <Button variant="solid" size="small">
            检查更新
          </Button>
          <Radio.Group
            size="small"
            value={theme}
            onChange={handleThemeChange}
            buttonStyle="solid"
          >
            <Radio.Button value="light">亮色主题</Radio.Button>
            <Radio.Button value="dark">黑色主题</Radio.Button>
            <Radio.Button value="system">跟随系统</Radio.Button>
          </Radio.Group>
        </Space>
      </Card>

      <Card title="日志" size="small">
        <Space>
          <div>
            <span>自动清空&nbsp;&nbsp;</span>
            <Tooltip
              placement="rightTop"
              title={"当日志数量达到设定值时即自动清空当前所有日志。"}
            >
              <QuestionCircleOutlined
                style={{ color: "#555555", fontSize: "16px" }}
              />
            </Tooltip>
            <InputNumber
              size="small"
              min={100}
              max={500}
              value={
                wechatConfig?.clearLog ? Number(wechatConfig.clearLog) : 500
              }
              onChange={onChangeClearLog}
            />
          </div>
          <Checkbox
            checked={wechatConfig?.logs || false}
            onChange={onChangeLogs}
            disabled
          >
            显示日志
          </Checkbox>
          <Tooltip placement="rightTop" title={"根据环境框架内置日志插件。"}>
            <QuestionCircleOutlined
              style={{ color: "#555555", fontSize: "16px" }}
            />
          </Tooltip>
        </Space>
      </Card>
      <Card size="small" title="微信">
        <Space direction="vertical" size={12} style={{ width: "100%" }}>
          <div className="flex items-center gap-2">
            <span className="text-right w-16">安装目录</span>
            <Tooltip
              placement="rightTop"
              title={
                "安装微信客户端的路径，末尾无需添加斜杠\\或/，举例：D:soft\\Weixin"
              }
            >
              <QuestionCircleOutlined
                style={{ color: "#555555", fontSize: "16px" }}
              />
            </Tooltip>
            <Input
              placeholder="请输入微信安装位置或获取..."
              size="small"
              className="flex-1"
              value={wechatConfig?.installationPath || ""}
              onChange={onChangeInstallPath}
            />
          </div>
          <div className="flex items-center gap-2">
            <span className="text-right w-16">缓存目录</span>
            <Tooltip
              placement="rightTop"
              title={
                "微信缓存路径，保存解密后的图片以及文件，末尾以斜杠\\结尾。举例：C:\\Users\\Administrator\\Documents\\NdogCache\\"
              }
            >
              <QuestionCircleOutlined
                style={{ color: "#555555", fontSize: "16px" }}
              />
            </Tooltip>
            <Input
              placeholder="请输入解密图片、文件等缓存位置或获取..."
              size="small"
              className="flex-1"
              value={wechatConfig?.cachePath || ""}
              onChange={onChangeCachePath}
            />
          </div>
          <div className="get-path-btn">
            <Button
              onClick={getWechatPath}
              variant="solid"
              size="small"
              className="w-full"
            >
              点我自动获取
            </Button>
          </div>
        </Space>
      </Card>
      <Card title="框架" size="small">
        <Space>
          <div>
            <span>微信版本&nbsp;&nbsp;</span>{" "}
            <Tooltip
              placement="rightTop"
              title={"建议选择最新支持的微信版本。"}
            >
              <QuestionCircleOutlined
                style={{ color: "#555555", fontSize: "16px" }}
              />
            </Tooltip>
            <Select
              value={wechatConfig?.version?.[0] || "4.12.17"}
              style={{ width: 100 }}
              onChange={handleChangeVersion}
              size="small"
              options={[{ value: "4.12.17", label: "4.12.17" }]}
            />
          </div>
          <div>
            <span>禁止更新&nbsp;&nbsp;</span>
            <Tooltip
              placement="rightTop"
              title={"建议开启，否则微信强制更新会导致框架不稳定。"}
            >
              <QuestionCircleOutlined
                style={{ color: "#555555", fontSize: "16px" }}
              />
            </Tooltip>
            <Select
              value={wechatConfig?.update || "1"}
              style={{ width: 100 }}
              onChange={handleChangeUpdate}
              size="small"
              options={[
                { value: "0", label: "关闭" },
                { value: "1", label: "开启" },
              ]}
            />
          </div>
          <div>
            <span>解密图片&nbsp;&nbsp;</span>
            <Tooltip
              placement="rightTop"
              title={
                "收到图片消息是否需要解密，如果无此应用场景，建议关闭，提高框架性能。"
              }
            >
              <QuestionCircleOutlined
                style={{ color: "#555555", fontSize: "16px" }}
              />
            </Tooltip>
            <Select
              value={wechatConfig?.decodePict || "0"}
              style={{ width: 100 }}
              onChange={handleChangeUpdate}
              size="small"
              options={[
                { value: "0", label: "关闭" },
                { value: "1", label: "开启" },
              ]}
            />
          </div>
          <div>
            <span>下载超时&nbsp;&nbsp;</span>
            <Tooltip
              placement="rightTop"
              title={
                "为确保成功解密图片，框架需先下载原图。在网络较慢时，请适当延长下载超时时间（秒），避免因超时导致下载失败。"
              }
            >
              <QuestionCircleOutlined
                style={{ color: "#555555", fontSize: "16px" }}
              />
            </Tooltip>
            <InputNumber
              style={{ width: 100 }}
              size="small"
              min={1}
              max={100}
              value={
                wechatConfig?.timeOut ? Number(wechatConfig.timeOut) / 1000 : 5
              }
              onChange={onChangeTimeOut} // 添加这一行
            />
          </div>

          <div>
            <span>忽略消息&nbsp;&nbsp;</span>
            <Tooltip
              placement="rightTop"
              title={
                "为避免处理微信登录瞬间的缓存消息，可设定在登录成功后的特定时长内，自动忽略收到的消息。为确保生效，建议此处设置一个较长的秒数。"
              }
            >
              <QuestionCircleOutlined
                style={{ color: "#555555", fontSize: "16px" }}
              />
            </Tooltip>
            <InputNumber
              style={{ width: 100 }}
              size="small"
              min={1}
              max={100}
              value={
                wechatConfig?.ignoreMsg ? Number(wechatConfig.ignoreMsg) : 5
              }
              onChange={onChangeIgnoreMsg}
            />
          </div>
        </Space>
        <Space>
          <div>
            <span>消息重发&nbsp;&nbsp;</span>
            <Tooltip
              placement="rightTop"
              title={
                "文本消息发送失败(转圈圈等情况)是否需要自动重新发送，框架每秒检查一次发送状态，连续检查到最多3次发送失败后，开始自动重发，若3次重发后仍失败，则停止并放弃发送。"
              }
            >
              <QuestionCircleOutlined
                style={{ color: "#555555", fontSize: "16px" }}
              />
            </Tooltip>
            <Select
              style={{ width: 100 }}
              value={wechatConfig?.resend || "1"}
              onChange={handleChangeResend}
              size="small"
              options={[
                { value: "0", label: "关闭" },
                { value: "1", label: "开启" },
              ]}
            />
          </div>
          <div>
            <span>监控群员&nbsp;&nbsp;</span>
            <Tooltip
              placement="rightTop"
              title={
                "若不需要监听群成员入群、退群消息，建议关闭，提升框架性能。"
              }
            >
              <QuestionCircleOutlined
                style={{ color: "#555555", fontSize: "16px" }}
              />
            </Tooltip>
            <Select
              style={{ width: 100 }}
              value={wechatConfig?.groupMemberEvent || "0"}
              onChange={handleChangeGroupMemberEvent}
              size="small"
              options={[
                { value: "0", label: "关闭" },
                { value: "1", label: "开启" },
              ]}
            />
          </div>
          <div>
            <span>语音文件&nbsp;&nbsp;</span>
            <Tooltip
              placement="rightTop"
              title={
                "根据语音ID获取SILK文件，无此应用场景建议关闭，提高框架性能。"
              }
            >
              <QuestionCircleOutlined
                style={{ color: "#555555", fontSize: "16px" }}
              />
            </Tooltip>
            <Select
              style={{ width: 100 }}
              value={wechatConfig?.hookSilk || "0"}
              onChange={handleChangeHookSilkEvent}
              size="small"
              options={[
                { value: "0", label: "关闭" },
                { value: "1", label: "开启" },
              ]}
            />
          </div>
          <div>
            <span>防封特征&nbsp;&nbsp;</span>
            <Tooltip
              placement="rightTop"
              title={
                "降低行为被微信安全机制识别与处置的风险，任何操作行为均需遵守微信官方相关规定。"
              }
            >
              <QuestionCircleOutlined
                style={{ color: "#555555", fontSize: "16px" }}
              />
            </Tooltip>
            <Select
              value={"1"}
              style={{ width: 100 }}
              size="small"
              options={[
                { value: "0", label: "关闭" },
                { value: "1", label: "开启" },
              ]}
            />
          </div>
          <div>
            <span>通信模式&nbsp;&nbsp;</span>
            <Tooltip
              placement="rightTop"
              title={
                "框架与微信之间的底层通信方式，暂定采用HTTP协议，通用性强，后期会考虑增加更高效的TCP协议。"
              }
            >
              <QuestionCircleOutlined
                style={{ color: "#555555", fontSize: "16px" }}
              />
            </Tooltip>
            <Input
              variant="filled"
              style={{ width: 100 }}
              size="small"
              value={"Https"}
            />
          </div>
        </Space>
      </Card>

      <div></div>
    </div>
  );
};

export default Settings;
