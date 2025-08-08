package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"time"
)

// MockHuifuClient 模拟汇付客户端实现
// 实际使用时需要根据 bspay-go-sdk 的具体API进行调整
type MockHuifuClient struct {
	sysID         string
	productID     string
	rsaPrivateKey *rsa.PrivateKey
	isProduction  bool
	httpClient    *http.Client
}

// NewMockHuifuClient 创建模拟客户端
func NewMockHuifuClient(config *ConfigRequest, isProduction bool) (*MockHuifuClient, error) {
	// 解析RSA私钥
	block, _ := pem.Decode([]byte(config.RSAPrivateKey))
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// 尝试PKCS8格式
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %v", err)
		}
		var ok bool
		privateKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("not an RSA private key")
		}
	}

	return &MockHuifuClient{
		sysID:         config.SysID,
		productID:     config.ProductID,
		rsaPrivateKey: privateKey,
		isProduction:  isProduction,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// CallAPI 调用汇付API（模拟版本）
func (c *MockHuifuClient) CallAPI(endpoint string, params map[string]interface{}) (map[string]interface{}, error) {
	fmt.Printf("\n=== CallAPI Start (Mock Mode) ===\n")
	fmt.Printf("Endpoint: %s\n", endpoint)
	fmt.Printf("Params: %+v\n", params)

	// 添加系统参数
	params["sys_id"] = c.sysID
	params["product_id"] = c.productID
	params["timestamp"] = time.Now().Format("20060102150405")

	// 生成签名（模拟）
	signature, err := c.generateSignature(params)
	if err != nil {
		fmt.Printf("Signature generation failed: %v\n", err)
		return nil, fmt.Errorf("failed to generate signature: %v", err)
	}
	params["sign"] = signature

	// 打印请求参数（用于调试）
	jsonData, _ := json.MarshalIndent(params, "", "  ")
	fmt.Printf("Request params (formatted):\n%s\n", string(jsonData))

	// 模拟不同API的响应
	var mockResult map[string]interface{}

	switch endpoint {
	case "/v2/merchant/busi/config", "/v2/merchant/wechat/config":
		// 模拟微信商户配置成功响应
		mockResult = map[string]interface{}{
			"resp_code": "00000",
			"resp_desc": "成功",
			"data": map[string]interface{}{
				"huifu_id":      params["huifu_id"],
				"wx_woa_app_id": params["wx_woa_app_id"],
				"wx_woa_path":   params["wx_woa_path"],
				"fee_type":      params["fee_type"],
				"config_status": "SUCCESS",
				"config_time":   time.Now().Format("2006-01-02 15:04:05"),
			},
		}

	case "/v2/merchant/busi/config/query":
		// 模拟微信商户配置查询响应
		mockResult = map[string]interface{}{
			"resp_code": "00000",
			"resp_desc": "成功",
			"data": map[string]interface{}{
				"huifu_id":      params["huifu_id"],
				"wx_woa_app_id": "wx1234567890abcdef",
				"wx_woa_path":   "pages/index/index",
				"fee_type":      "01",
				"config_status": "ACTIVE",
				"config_time":   "2024-01-01 10:00:00",
				"update_time":   time.Now().Format("2006-01-02 15:04:05"),
			},
		}

	case "/v2/merchant/basicdata/query":
		// 模拟商户信息查询响应
		mockResult = map[string]interface{}{
			"resp_code": "00000",
			"resp_desc": "成功",
			"data": map[string]interface{}{
				"sys_id":      c.sysID,
				"product_id":  c.productID,
				"status":      "ACTIVE",
				"create_time": "2024-01-01 10:00:00",
			},
		}

	default:
		// 默认成功响应
		mockResult = map[string]interface{}{
			"resp_code": "00000",
			"resp_desc": "成功",
			"data": map[string]interface{}{
				"message": fmt.Sprintf("Mock response for endpoint: %s", endpoint),
			},
		}
	}

	fmt.Printf("Mock response: %+v\n", mockResult)
	fmt.Printf("=== CallAPI End (Mock Mode) ===\n\n")

	return mockResult, nil
}

// generateSignature 生成RSA签名
func (c *MockHuifuClient) generateSignature(params map[string]interface{}) (string, error) {
	// 将参数转换为JSON字符串
	jsonData, err := json.Marshal(params)
	if err != nil {
		return "", err
	}

	// 计算SHA256哈希
	hashed := sha256.Sum256(jsonData)

	// RSA签名
	signature, err := rsa.SignPKCS1v15(rand.Reader, c.rsaPrivateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}

	// Base64编码
	return base64.StdEncoding.EncodeToString(signature), nil
}

// 特定的API方法实现

// ConfigureWeChatMerchant 配置微信商户
func (c *MockHuifuClient) ConfigureWeChatMerchant(huifuID, wxAppID, wxPath string) (map[string]interface{}, error) {
	params := map[string]interface{}{
		"huifu_id":      huifuID,
		"wx_woa_app_id": wxAppID,
		"wx_woa_path":   wxPath,
		"req_seq_id":    generateReqSeqID(),
		"req_date":      time.Now().Format("20060102"),
	}

	// 调用微信商户配置API
	// 这里的endpoint需要根据实际API文档调整
	return c.CallAPI("/v2/merchant/wechat/config", params)
}

// QueryMerchantInfo 查询商户信息
func (c *MockHuifuClient) QueryMerchantInfo(huifuID string) (map[string]interface{}, error) {
	params := map[string]interface{}{
		"huifu_id":   huifuID,
		"req_seq_id": generateReqSeqID(),
		"req_date":   time.Now().Format("20060102"),
	}

	return c.CallAPI("/v2/merchant/basicdata/query", params)
}

// generateReqSeqID 生成请求序列号
func generateReqSeqID() string {
	return fmt.Sprintf("%d%06d", time.Now().Unix(), time.Now().Nanosecond()/1000)
}
