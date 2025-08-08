package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// WeChatConfigRequest 微信商户配置请求
type WeChatConfigRequest struct {
	ReqSeqId    string                 `json:"req_seq_id"`
	ReqDate     string                 `json:"req_date"`
	HuifuId     string                 `json:"huifu_id"`
	ExtendInfos map[string]interface{} `json:"extend_infos"`
}

// ConfigManager 管理动态配置
type ConfigManager struct {
	mu         sync.RWMutex
	configs    map[string]*ConfigRequest
	sdkClients map[string]HuifuClient
	configFile string
}

// NewConfigManager 创建配置管理器
func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		configs:    make(map[string]*ConfigRequest),
		sdkClients: make(map[string]HuifuClient),
		configFile: "./temp_config.json",
	}
}

// SaveConfig 保存配置并初始化SDK客户端
func (cm *ConfigManager) SaveConfig(config *ConfigRequest) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 生成临时配置文件用于SDK初始化
	configData := map[string]interface{}{
		"default": map[string]interface{}{
			"sys_id":               config.SysID,
			"product_id":           config.ProductID,
			"rsa_private_key":      config.RSAPrivateKey,
			"rsa_huifu_public_key": "", // 需要从汇付获取
		},
	}

	// 写入临时配置文件
	jsonData, err := json.MarshalIndent(configData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	err = os.WriteFile(cm.configFile, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	// 初始化SDK客户端
	isProd := config.Environment == "production"

	// 优先使用真实的SDK客户端
	var sdkClient HuifuClient

	sdkClient, err = NewRealHuifuClient(config, isProd)
	if err != nil {
		// 如果真实客户端初始化失败，尝试使用模拟客户端
		log.Printf("Failed to initialize real SDK client: %v, falling back to mock client", err)
		sdkClient, err = NewMockHuifuClient(config, isProd)
		if err != nil {
			return fmt.Errorf("failed to initialize any SDK client: %v", err)
		}
	}

	// 存储配置和客户端
	configKey := config.SysID
	cm.configs[configKey] = config
	cm.sdkClients[configKey] = sdkClient

	return nil
}

// GetSDKClient 获取SDK客户端
func (cm *ConfigManager) GetSDKClient(sysID string) (HuifuClient, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	client, exists := cm.sdkClients[sysID]
	if !exists {
		return nil, fmt.Errorf("SDK client not found for sys_id: %s", sysID)
	}
	return client, nil
}

// DeleteConfig 删除配置
func (cm *ConfigManager) DeleteConfig(sysID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 检查配置是否存在
	if _, exists := cm.configs[sysID]; !exists {
		return fmt.Errorf("configuration not found for sys_id: %s", sysID)
	}

	// 清理SDK客户端
	if client, exists := cm.sdkClients[sysID]; exists {
		// 如果是真实客户端，清理临时文件
		if realClient, ok := client.(*RealHuifuClient); ok {
			realClient.Cleanup()
		}
		delete(cm.sdkClients, sysID)
	}

	// 删除配置
	delete(cm.configs, sysID)

	// 清理临时配置文件
	configPath := fmt.Sprintf("./config_%s.json", sysID)
	os.Remove(configPath)

	return nil
}

var configManager = NewConfigManager()

func main() {
	r := gin.Default()

	// 配置CORS
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	r.Use(cors.New(config))

	// 静态文件服务
	r.Static("/static", "./static")
	r.StaticFile("/", "./static/index.html")

	// API路由
	api := r.Group("/api")
	{
		// 保存配置
		api.POST("/config", saveConfig)

		// 删除配置
		api.DELETE("/config/:sys_id", deleteConfig)

		// 测试配置
		api.POST("/test-config", testConfig)

		// 配置微信商户
		api.POST("/wechat-config", configureWeChatMerchant)

		// 查询微信商户配置
		api.POST("/wechat-config-query", queryWeChatConfig)

		// 获取配置列表
		api.GET("/configs", getConfigs)

		// 生成测试密钥
		api.GET("/generate-test-key", generateTestKey)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // 默认端口
	}
	log.Println("Server starting on :" + port + "...")
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// saveConfig 保存配置处理函数
func saveConfig(c *gin.Context) {
	var config ConfigRequest
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	// 设置默认环境
	if config.Environment == "" {
		config.Environment = "test"
	}

	if len(config.RSAPrivateKey) > 0 {
		// 如果缺少 BEGIN 和 END 标记，添加它们
		if !strings.Contains(config.RSAPrivateKey, "BEGIN RSA PRIVATE KEY") {
			config.RSAPrivateKey = "-----BEGIN RSA PRIVATE KEY-----\n" + config.RSAPrivateKey + "\n-----END RSA PRIVATE KEY-----"
		}
	}

	// 保存配置
	if err := configManager.SaveConfig(&config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to save configuration",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Configuration saved successfully",
		"sys_id":  config.SysID,
	})
}

// testConfig 测试配置是否有效
func testConfig(c *gin.Context) {
	var req struct {
		SysID string `json:"sys_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	// 获取SDK客户端
	client, err := configManager.GetSDKClient(req.SysID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Configuration not found",
			"details": err.Error(),
		})
		return
	}

	// 执行一个简单的测试调用
	testParams := map[string]interface{}{
		"test": true,
	}

	_, err = client.CallAPI("/v2/merchant/basicdata/query", testParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Configuration test failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Configuration is valid",
		"status":  "success",
	})
}

// configureWeChatMerchant 配置微信商户
func configureWeChatMerchant(c *gin.Context) {
	log.Println("=== configureWeChatMerchant Start ===")

	var req struct {
		SysID       string                 `json:"sys_id" binding:"required"`
		HuifuID     string                 `json:"huifu_id" binding:"required"`
		WxWoaAppID  string                 `json:"wx_woa_app_id" binding:"required"`
		WxWoaPath   string                 `json:"wx_woa_path" binding:"required"`
		FeeType     string                 `json:"fee_type" binding:"required"`
		ExtendInfos map[string]interface{} `json:"extend_infos"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Request binding failed: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	log.Printf("Request received: %+v\n", req)

	// 获取SDK客户端
	client, err := configManager.GetSDKClient(req.SysID)
	if err != nil {
		log.Printf("SDK client not found: %v\n", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Configuration not found",
			"details": err.Error(),
		})
		return
	}

	log.Printf("SDK client retrieved for sys_id: %s\n", req.SysID)

	// 构建扩展信息
	if req.ExtendInfos == nil {
		req.ExtendInfos = make(map[string]interface{})
	}
	req.ExtendInfos["wx_woa_app_id"] = req.WxWoaAppID
	req.ExtendInfos["wx_woa_path"] = req.WxWoaPath
	req.ExtendInfos["fee_type"] = req.FeeType

	// 构建API参数
	apiParams := map[string]interface{}{
		"huifu_id":      req.HuifuID,
		"wx_woa_app_id": req.WxWoaAppID,
		"wx_woa_path":   req.WxWoaPath,
		"fee_type":      req.FeeType,
	}

	// 合并扩展信息
	for k, v := range req.ExtendInfos {
		apiParams[k] = v
	}

	log.Printf("API params built: %+v\n", apiParams)

	// 调用API
	log.Println("Calling client.CallAPI...")
	result, err := client.CallAPI("/v2/merchant/busi/config", apiParams)
	if err != nil {
		log.Printf("CallAPI failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to configure WeChat merchant",
			"details": err.Error(),
		})
		return
	}

	log.Printf("CallAPI successful, result: %+v\n", result)

	c.JSON(http.StatusOK, gin.H{
		"message":   result["data"],
		"huifu_id":  req.HuifuID,
		"wx_app_id": req.WxWoaAppID,
	})

	log.Println("=== configureWeChatMerchant End ===")
}

// queryWeChatConfig 查询微信商户配置
func queryWeChatConfig(c *gin.Context) {
	log.Println("=== queryWeChatConfig Start ===")

	var req struct {
		SysID   string `json:"sys_id" binding:"required"`
		HuifuID string `json:"huifu_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Request binding failed: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	log.Printf("Request received: %+v\n", req)

	// 获取SDK客户端
	client, err := configManager.GetSDKClient(req.SysID)
	if err != nil {
		log.Printf("SDK client not found: %v\n", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Configuration not found",
			"details": err.Error(),
		})
		return
	}

	log.Printf("SDK client retrieved for sys_id: %s\n", req.SysID)

	// 构建API参数 - 查询只需要huifu_id
	apiParams := map[string]interface{}{
		"huifu_id": req.HuifuID,
	}

	log.Printf("API params built: %+v\n", apiParams)

	// 调用API
	log.Println("Calling client.CallAPI for query...")
	result, err := client.CallAPI("/v2/merchant/busi/config/query", apiParams)
	if err != nil {
		log.Printf("CallAPI failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to query WeChat merchant config",
			"details": err.Error(),
		})
		return
	}

	log.Printf("CallAPI successful, result: %+v\n", result)

	c.JSON(http.StatusOK, gin.H{
		"message":  result["data"],
		"huifu_id": req.HuifuID,
	})

	log.Println("=== queryWeChatConfig End ===")
}

// deleteConfig 删除配置处理函数
func deleteConfig(c *gin.Context) {
	sysID := c.Param("sys_id")

	if sysID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "sys_id is required",
		})
		return
	}

	if err := configManager.DeleteConfig(sysID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Configuration not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Configuration deleted successfully",
		"sys_id":  sysID,
	})
}

// getConfigs 获取所有配置
func getConfigs(c *gin.Context) {
	configManager.mu.RLock()
	defer configManager.mu.RUnlock()

	configs := []map[string]string{}
	for sysID, config := range configManager.configs {
		configs = append(configs, map[string]string{
			"sys_id":      sysID,
			"product_id":  config.ProductID,
			"environment": config.Environment,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"configs": configs,
		"count":   len(configs),
	})
}

// generateTestKey 生成测试RSA密钥
func generateTestKey(c *gin.Context) {
	// 生成2048位的RSA密钥对
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate RSA key",
			"details": err.Error(),
		})
		return
	}

	// 将私钥编码为PEM格式
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	privateKeyString := string(pem.EncodeToMemory(privateKeyPEM))

	// 生成公钥
	publicKeyPKIX, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to marshal public key",
			"details": err.Error(),
		})
		return
	}

	publicKeyPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyPKIX,
	}

	publicKeyString := string(pem.EncodeToMemory(publicKeyPEM))

	c.JSON(http.StatusOK, gin.H{
		"private_key": privateKeyString,
		"public_key":  publicKeyString,
		"message":     "Test RSA key pair generated successfully",
	})
}
