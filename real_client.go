package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/huifurepo/bspay-go-sdk/BsPaySdk"
	"github.com/huifurepo/bspay-go-sdk/ut/tool"
)

// RealHuifuClient 真实的汇付SDK客户端
type RealHuifuClient struct {
	sdk          *BsPaySdk.BsPay
	config       *ConfigRequest
	isProduction bool
}

// NewRealHuifuClient 创建真实的SDK客户端
func NewRealHuifuClient(config *ConfigRequest, isProduction bool) (*RealHuifuClient, error) {
	fmt.Printf("Creating real SDK client for sys_id: %s\n", config.SysID)

	// 创建临时配置文件
	configPath := fmt.Sprintf("./config_%s.json", config.SysID)
	fmt.Printf("Config file path: %s\n", configPath)

	// 构建配置数据 - 需要符合SDK期望的格式
	// 根据环境选择汇付公钥
	var huifuPublicKey string
	if isProduction {
		// 生产环境的汇付公钥（需要从汇付获取）
		huifuPublicKey = `MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAkMX8p3GyMw3gk6x72h20NOk3L9+Nn9mOVP6+YoBwCe7Zs4QmYrA/etFRZw2TQrSc51wgtCkJi1/x8Wl7maPL1uH2+77JFlPv7H/F4Lr2I2LXgnllg6PtwOSw/qvGYInVVB4kL85VQl0/8ObyxBUdJ43I0z/u8hJb2gwujSudOGizbeqQXAYrwcNy+e+cjodpPy9unpJjBfa4Wz2eVLLvUYYKZKdRn6pZR2cQsMBvL30K4cFlZqlJ9iP2hTG3gaiZJ9JrjTigwki0g9pbTDXiPACfuF1nOeObvLD22zBbgn1kwgfsqoG67z7g84u2jvfUFCzX1JRgd0xfNorTRkS2RQIDAQAB`
	} else {
		// 测试环境的汇付公钥（这是SDK demo中的测试公钥）
		huifuPublicKey = `MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAkMX8p3GyMw3gk6x72h20NOk3L9+Nn9mOVP6+YoBwCe7Zs4QmYrA/etFRZw2TQrSc51wgtCkJi1/x8Wl7maPL1uH2+77JFlPv7H/F4Lr2I2LXgnllg6PtwOSw/qvGYInVVB4kL85VQl0/8ObyxBUdJ43I0z/u8hJb2gwujSudOGizbeqQXAYrwcNy+e+cjodpPy9unpJjBfa4Wz2eVLLvUYYKZKdRn6pZR2cQsMBvL30K4cFlZqlJ9iP2hTG3gaiZJ9JrjTigwki0g9pbTDXiPACfuF1nOeObvLD22zBbgn1kwgfsqoG67z7g84u2jvfUFCzX1JRgd0xfNorTRkS2RQIDAQAB`
	}

	// 处理RSA私钥格式 - SDK期望纯内容，不带BEGIN/END标记
	privateKey := config.RSAPrivateKey
	privateKey = strings.ReplaceAll(privateKey, "-----BEGIN RSA PRIVATE KEY-----", "")
	privateKey = strings.ReplaceAll(privateKey, "-----END RSA PRIVATE KEY-----", "")
	privateKey = strings.ReplaceAll(privateKey, "\n", "")
	privateKey = strings.TrimSpace(privateKey)

	// SDK期望的配置文件格式是平铺的，不是嵌套的
	configData := map[string]interface{}{
		"sys_id":                config.SysID,
		"product_id":            config.ProductID,
		"rsa_merch_private_key": privateKey,
		"rsa_huifu_public_key":  huifuPublicKey,
	}

	// 写入配置文件
	jsonData, err := json.MarshalIndent(configData, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %v", err)
	}

	fmt.Printf("Writing config to file: %s\n", configPath)
	if err := os.WriteFile(configPath, jsonData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write config file: %v", err)
	}
	fmt.Printf("Config file written, size: %d bytes\n", len(jsonData))

	// 初始化SDK
	fmt.Printf("Initializing SDK with production=%v, config=%s\n", isProduction, configPath)
	sdk, err := BsPaySdk.NewBsPay(isProduction, configPath)
	if err != nil {
		// 清理配置文件
		os.Remove(configPath)
		return nil, fmt.Errorf("failed to initialize SDK: %v", err)
	}
	fmt.Printf("SDK initialized successfully\n")

	return &RealHuifuClient{
		sdk:          sdk,
		config:       config,
		isProduction: isProduction,
	}, nil
}

// CallAPI 调用汇付API
func (c *RealHuifuClient) CallAPI(endpoint string, params map[string]interface{}) (map[string]interface{}, error) {
	// 根据不同的endpoint调用不同的SDK方法
	switch endpoint {
	case "/v2/merchant/busi/config":
		return c.configureWeChatMerchant(params)
	case "/v2/merchant/busi/config/query":
		return c.queryMerchantConfig(params)
	default:
		return nil, fmt.Errorf("unsupported endpoint: %s", endpoint)
	}
}

// configureWeChatMerchant 配置微信商户
func (c *RealHuifuClient) configureWeChatMerchant(params map[string]interface{}) (map[string]interface{}, error) {
	// 构建微信商户配置请求
	req := BsPaySdk.V2MerchantBusiConfigRequest{
		ReqSeqId: tool.GetReqSeqId(),
		ReqDate:  tool.GetCurrentDate(),
		HuifuId:  params["huifu_id"].(string),
	}

	// 添加扩展信息
	req.ExtendInfos = map[string]interface{}{
		"wx_woa_app_id": params["wx_woa_app_id"],
		"wx_woa_path":   params["wx_woa_path"],
		"fee_type":      params["fee_type"],
	}

	// 如果有额外的参数，也添加进去
	for k, v := range params {
		if k != "huifu_id" && k != "wx_woa_app_id" && k != "wx_woa_path" && k != "fee_type" {
			if req.ExtendInfos == nil {
				req.ExtendInfos = make(map[string]interface{})
			}
			req.ExtendInfos[k] = v
		}
	}

	// 调用SDK方法 - 微信关注配置
	result, err := c.sdk.V2MerchantBusiConfigRequest(req)
	if err != nil {
		return nil, fmt.Errorf("WeChat config SDK call failed: %v", err)
	}

	return result, nil
}

// queryMerchantConfig 查询商户配置
func (c *RealHuifuClient) queryMerchantConfig(params map[string]interface{}) (map[string]interface{}, error) {
	// 构建查询请求
	huifuID := c.config.SysID
	if hid, ok := params["huifu_id"].(string); ok && hid != "" {
		huifuID = hid
	}

	req := BsPaySdk.V2MerchantBusiConfigQueryRequest{
		ReqSeqId: tool.GetReqSeqId(),
		ReqDate:  tool.GetCurrentDate(),
		HuifuId:  huifuID,
	}

	// 调用SDK
	result, err := c.sdk.V2MerchantBusiConfigQueryRequest(req)
	if err != nil {
		return nil, fmt.Errorf("merchant query SDK call failed: %v", err)
	}

	return result, nil
}

// 获取当前时间戳
func getCurrentTimestamp() string {
	return time.Now().Format("20060102150405")
}

// Cleanup 清理临时文件
func (c *RealHuifuClient) Cleanup() {
	configPath := fmt.Sprintf("./config_%s.json", c.config.SysID)
	os.Remove(configPath)
}
