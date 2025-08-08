# SDK Integration Guide
# SDK集成指南

## 当前实现状态

由于 `bspay-go-sdk` v1.0.20 是一个可执行程序而不是可导入的包，当前系统使用了一个**模拟客户端**来演示功能。这个模拟客户端实现了与真实SDK相同的接口，允许系统正常运行。

## 架构设计

```
┌─────────────────┐
│   Web界面       │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Gin API Server │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ HuifuClient接口 │ ← 抽象层
└────────┬────────┘
         │
    ┌────┴────┐
    ▼         ▼
┌─────────┐ ┌─────────┐
│MockClient│ │RealSDK │
└─────────┘ └─────────┘
```

## 文件说明

1. **main.go** - 主服务器程序
   - 定义了 `HuifuClient` 接口
   - 管理配置和客户端实例

2. **huifu_client.go** - 模拟客户端实现
   - `MockHuifuClient` - 模拟SDK功能
   - 实现了签名和API调用逻辑

3. **sdk_adapter.go** - SDK适配器（备选方案）
   - 通过子进程方式调用SDK
   - 适用于SDK作为独立程序的情况

## 集成真实SDK的方法

### 方法1: 如果SDK可以作为包导入

如果未来版本的 `bspay-go-sdk` 支持作为包导入，创建一个新的实现：

```go
// real_client.go
package main

import (
    BsPaySdk "github.com/huifurepo/bspay-go-sdk/lib" // 假设的包路径
)

type RealHuifuClient struct {
    sdk *BsPaySdk.BsPay
}

func NewRealHuifuClient(config *ConfigRequest, isProd bool) (*RealHuifuClient, error) {
    // 创建配置文件
    configPath := createConfigFile(config)
    
    // 初始化SDK
    sdk, err := BsPaySdk.NewBsPay(isProd, configPath)
    if err != nil {
        return nil, err
    }
    
    return &RealHuifuClient{sdk: sdk}, nil
}

func (c *RealHuifuClient) CallAPI(endpoint string, params map[string]interface{}) (map[string]interface{}, error) {
    // 根据endpoint调用相应的SDK方法
    switch endpoint {
    case "/v2/merchant/wechat/config":
        return c.configureWechat(params)
    case "/v2/merchant/basicdata/query":
        return c.queryMerchant(params)
    default:
        return nil, fmt.Errorf("unsupported endpoint: %s", endpoint)
    }
}
```

### 方法2: 使用HTTP代理模式

如果SDK只能作为独立服务运行，可以创建一个代理服务：

```go
// proxy_client.go
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
)

type ProxyHuifuClient struct {
    sdkServiceURL string
    httpClient    *http.Client
}

func NewProxyHuifuClient(config *ConfigRequest) (*ProxyHuifuClient, error) {
    // 启动SDK服务（假设SDK提供HTTP接口）
    sdkServiceURL := startSDKService(config)
    
    return &ProxyHuifuClient{
        sdkServiceURL: sdkServiceURL,
        httpClient:    &http.Client{Timeout: 30 * time.Second},
    }, nil
}

func (c *ProxyHuifuClient) CallAPI(endpoint string, params map[string]interface{}) (map[string]interface{}, error) {
    // 转发请求到SDK服务
    url := fmt.Sprintf("%s%s", c.sdkServiceURL, endpoint)
    
    jsonData, _ := json.Marshal(params)
    req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    return result, nil
}
```

### 方法3: 使用命令行调用

如果SDK提供命令行接口：

```go
// cli_client.go
package main

import (
    "encoding/json"
    "os/exec"
)

type CLIHuifuClient struct {
    configPath string
}

func NewCLIHuifuClient(config *ConfigRequest) (*CLIHuifuClient, error) {
    configPath := createConfigFile(config)
    return &CLIHuifuClient{configPath: configPath}, nil
}

func (c *CLIHuifuClient) CallAPI(endpoint string, params map[string]interface{}) (map[string]interface{}, error) {
    // 构建命令行参数
    jsonParams, _ := json.Marshal(params)
    
    cmd := exec.Command("bspay-cli", 
        "--config", c.configPath,
        "--endpoint", endpoint,
        "--params", string(jsonParams),
    )
    
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }
    
    var result map[string]interface{}
    json.Unmarshal(output, &result)
    return result, nil
}
```

## 切换客户端实现

在 `main.go` 的 `SaveConfig` 方法中切换实现：

```go
func (cm *ConfigManager) SaveConfig(config *ConfigRequest) error {
    // ... 其他代码 ...
    
    // 根据环境变量选择客户端实现
    var sdkClient HuifuClient
    var err error
    
    switch os.Getenv("SDK_MODE") {
    case "real":
        sdkClient, err = NewRealHuifuClient(config, isProd)
    case "proxy":
        sdkClient, err = NewProxyHuifuClient(config)
    case "cli":
        sdkClient, err = NewCLIHuifuClient(config)
    default:
        sdkClient, err = NewMockHuifuClient(config, isProd)
    }
    
    if err != nil {
        return fmt.Errorf("failed to initialize SDK client: %v", err)
    }
    
    // ... 其他代码 ...
}
```

## 环境配置

在 `.env` 文件中配置SDK模式：

```bash
# SDK集成模式
SDK_MODE=mock  # mock, real, proxy, cli

# SDK相关配置
SDK_SERVICE_URL=http://localhost:9090
SDK_CLI_PATH=/usr/local/bin/bspay-cli
```

## API映射

根据汇付API文档，以下是常用的API映射：

| 功能 | SDK方法 | API端点 |
|------|---------|---------|
| 微信商户配置 | V2MerchantWechatConfig | /v2/merchant/wechat/config |
| 商户信息查询 | V2MerchantBasicdataQuery | /v2/merchant/basicdata/query |
| 支付下单 | V2TradePaymentJspay | /v2/trade/payment/jspay |
| 订单查询 | V2TradeQueryOrder | /v2/trade/query/order |
| 退款申请 | V2TradeRefund | /v2/trade/refund |

## 测试集成

1. **单元测试**
```go
func TestHuifuClient(t *testing.T) {
    config := &ConfigRequest{
        SysID:         "test_sys",
        ProductID:     "test_product",
        RSAPrivateKey: testPrivateKey,
    }
    
    client, err := NewMockHuifuClient(config, false)
    assert.NoError(t, err)
    
    result, err := client.CallAPI("/test", map[string]interface{}{})
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

2. **集成测试**
```bash
# 启动服务器
go run .

# 测试API
curl -X POST http://localhost:8080/api/config \
  -H "Content-Type: application/json" \
  -d '{
    "sys_id": "test",
    "product_id": "test",
    "rsa_private_key": "-----BEGIN RSA PRIVATE KEY-----..."
  }'
```

## 故障排查

### 常见问题

1. **SDK导入失败**
   - 确认SDK版本和导入路径
   - 检查go.mod依赖配置

2. **签名验证失败**
   - 确认RSA私钥格式正确
   - 检查参数排序和编码

3. **API调用失败**
   - 验证API端点路径
   - 检查必需参数是否完整

## 下一步

1. 联系汇付技术支持获取：
   - SDK的正确集成方式
   - RSA公钥
   - API文档和示例代码

2. 根据实际SDK接口调整客户端实现

3. 添加完整的错误处理和日志记录

4. 实现配置的持久化存储（数据库）

5. 添加API调用的监控和告警