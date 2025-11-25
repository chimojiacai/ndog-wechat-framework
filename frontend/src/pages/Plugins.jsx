import {
  Card,
  Row,
  Col,
  Button,
  Empty,
  Tag,
  App,
  Upload,
  Popconfirm,
  Pagination,
  Space,
  Avatar,
  Spin,
} from "antd";
import {
  UploadOutlined,
  DeleteOutlined,
  LoadingOutlined,
  FolderOpenOutlined,
} from "@ant-design/icons";
import { useState, useEffect } from "react";
import {
  ScanPlugins,
  OpenPlugin,
  RefreshPlugins,
  UninstallPlugin,
} from "../../bindings/github.com/naidog/wechat-framework/service/plugin/pluginservice";

const Plugins = () => {
  const [plugins, setPlugins] = useState([]);
  const [loading, setLoading] = useState(true);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(4);
  const { message } = App.useApp();

  useEffect(() => {
    loadPlugins();
  }, []);

  const loadPlugins = async () => {
    try {
      const data = await ScanPlugins();
      setPlugins(data || []);
    } catch (error) {
      console.error("加载插件失败:", error);
    } finally {
      setLoading(false);
    }
  };

  const refreshPlugins = async () => {
    setLoading(true);
    try {
      const data = await RefreshPlugins();
      setPlugins(data || []);
      message.success(`刷新成功！共 ${data?.length || 0} 个插件`);
    } catch (error) {
      console.error("刷新插件失败:", error);
      message.error("刷新失败");
    } finally {
      setLoading(false);
    }
  };

  const uploadProps = {
    name: "file",
    accept: ".dog",
    showUploadList: false,
    beforeUpload: async (file) => {
      // 检查文件后缀
      if (!file.name.endsWith(".dog")) {
        message.error("请上传 .dog 格式的插件文件");
        return false;
      }

      // 检查文件大小（限制 50MB）
      const maxSize = 50 * 1024 * 1024; // 50MB
      if (file.size > maxSize) {
        message.error(`文件太大，请上传小于 50MB 的插件（当前: ${(file.size / 1024 / 1024).toFixed(2)}MB）`);
        return false;
      }

      const loadingMsg = message.loading(`正在上传 ${file.name} (${(file.size / 1024 / 1024).toFixed(2)}MB)...`, 0);

      try {
        // 使用 HTTP 上传
        const formData = new FormData();
        formData.append('file', file);

        const response = await fetch('http://localhost:9001/api/plugin/upload', {
          method: 'POST',
          body: formData,
        });

        const result = await response.json();
        
        loadingMsg(); // 关闭 loading

        if (result.code === 200) {
          message.success(`上传成功: ${file.name}，正在识别插件...`);
          
          // 延迟一下再刷新，等待文件解压完成
          setTimeout(() => {
            refreshPlugins();
          }, 1000);
        } else {
          message.error(`上传失败: ${result.msg}`);
        }
      } catch (error) {
        loadingMsg(); // 关闭 loading
        console.error("上传失败:", error);
        message.error(`上传失败: ${error.message}`);
      }

      return false; // 阻止默认上传行为
    },
  };

  const openPlugin = async (plugin) => {
    try {
      // 调用后端服务在新窗口中打开插件
      await OpenPlugin(plugin.metadata.id);
      message.success(`正在打开插件: ${plugin.metadata.name}`);
    } catch (error) {
      console.error("打开插件失败:", error);
      message.error(`打开失败: ${error}`);
    }
  };

  const uninstallPlugin = async (plugin) => {
    try {
      await UninstallPlugin(plugin.metadata.id);
      message.success(`${plugin.metadata.name}插件已卸载`);
      // 刷新列表
      refreshPlugins();
    } catch (error) {
      console.error("卸载插件失败:", error);
      message.error(`卸载失败: ${error}`);
    }
  };

  // 分页逻辑
  const startIndex = (currentPage - 1) * pageSize;
  const endIndex = startIndex + pageSize;
  const currentPlugins = plugins.slice(startIndex, endIndex);

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
          <Upload {...uploadProps}>
            <Button variant="solid" size="small">
              上传插件
            </Button>
          </Upload>
        </Space>
        <Button
          onClick={refreshPlugins}
          loading={loading}
          size="large"
          variant="solid"
          size="small"
        >
          刷新插件
        </Button>
      </Card>

      {loading ? (
        <span></span>
      ) : plugins.length === 0 ? (
        <div style={{ padding: "60px 0" }}>
          {/* <Empty description="暂无插件" /> */}
          <Empty description="No plugin" image={Empty.PRESENTED_IMAGE_SIMPLE} />
        </div>
      ) : (
        <>
          <Row gutter={[16, 16]}>
            {currentPlugins.map((plugin) => (
              <Col key={plugin.metadata.id} xs={24} sm={12} lg={6}>
                <Card
                  size="small"
                  style={{ height: "184px" }}
                  styles={{
                    body: {
                      height: "100%",
                      display: "flex",
                      flexDirection: "column",
                    },
                  }}
                >
                  {/* 插件图标和版本 */}
                  <div
                    style={{
                      display: "flex",
                      alignItems: "flex-start",
                      marginBottom: "12px",
                    }}
                  >
                    <Avatar
                      shape="square"
                      size={48}
                      src={plugin.iconUrl}
                      icon={<FolderOpenOutlined />}
                      style={{ marginRight: "12px", flexShrink: 0 }}
                    />
                    <div style={{ flex: 1, minWidth: 0 }}>
                      <div
                        style={{
                          fontWeight: 600,
                          marginBottom: "4px",
                          overflow: "hidden",
                          textOverflow: "ellipsis",
                          whiteSpace: "nowrap",
                        }}
                      >
                        {plugin.metadata.name}
                      </div>
                      <Tag>v{plugin.metadata.version}</Tag>
                    </div>
                  </div>

                  {/* 插件描述 */}
                  <div
                    style={{
                      fontSize: "12px",
                      overflow: "hidden",
                      textOverflow: "ellipsis",
                      color: "#606266",
                      whiteSpace: "nowrap",
                      flex: 1,
                      marginBottom: "1px",
                    }}
                  >
                    {plugin.metadata.description || "暂无描述"}
                  </div>

                  {/* 作者和时间 */}
                  <div
                    style={{
                      marginBottom: "12px",
                      fontSize: "12px",
                      color: "#606266",
                    }}
                  >
                    {plugin.metadata.author && (
                      <div>
                        {plugin.metadata.author}：
                        {new Date()
                          .toLocaleString("zh-CN", {
                            year: "numeric",
                            month: "2-digit",
                            day: "2-digit",
                            hour: "2-digit",
                            minute: "2-digit",
                            hour12: false,
                          })
                          .replace(/\//g, "-")}
                      </div>
                    )}
                  </div>

                  {/* 操作按钮 */}
                  <div
                    style={{ display: "flex", gap: "8px", marginTop: "auto" }}
                  >
                    <Button
                      variant="solid"
                      size="small"
                      style={{ flex: 1 }}
                      onClick={() => openPlugin(plugin)}
                    >
                      运行插件
                    </Button>
                    <Popconfirm
                      description={`确定要卸载「${plugin.metadata.name}」插件吗？`}
                      onConfirm={() => uninstallPlugin(plugin)}
                      okText="确定"
                      cancelText="取消"
                    >
                      <Button variant="solid" size="small" danger>
                        卸载插件
                      </Button>
                    </Popconfirm>
                  </div>
                </Card>
              </Col>
            ))}
          </Row>

          {/* 分页 */}
          {plugins.length > pageSize && (
            <div
              style={{
                display: "flex",
                justifyContent: "center",
              }}
            >
              <Pagination
                current={currentPage}
                pageSize={pageSize}
                total={plugins.length}
                onChange={(page, size) => {
                  setCurrentPage(page);
                  setPageSize(size);
                }}
                showTotal={(total) => `共 ${total} 个插件`}
              />
            </div>
          )}
        </>
      )}
    </div>
  );
};

export default Plugins;
