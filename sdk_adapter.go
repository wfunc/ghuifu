package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// BsPayAdapter 汇付SDK适配器
// 通过子进程调用方式集成 bspay-go-sdk
type BsPayAdapter struct {
	configPath   string
	isProduction bool
}

// NewBsPayAdapter 创建SDK适配器
func NewBsPayAdapter(config *ConfigRequest) (*BsPayAdapter, error) {
	// 创建临时配置文件
	tempDir := os.TempDir()
	configPath := filepath.Join(tempDir, fmt.Sprintf("huifu_config_%s.json", config.SysID))
	
	// 构建配置数据
	configData := map[string]interface{}{
		"default": map[string]interface{}{
			"sys_id":               config.SysID,
			"product_id":           config.ProductID,
			"rsa_private_key":      config.RSAPrivateKey,
			"rsa_huifu_public_key": "", // 需要从汇付获取
		},
	}
	
	// 写入配置文件
	jsonData, err := json.MarshalIndent(configData, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %v", err)
	}
	
	if err := os.WriteFile(configPath, jsonData, 0600); err != nil {
		return nil, fmt.Errorf("failed to write config file: %v", err)
	}
	
	return &BsPayAdapter{
		configPath:   configPath,
		isProduction: config.Environment == "production",
	}, nil
}

// CallAPI 调用API（通过子进程方式）
func (a *BsPayAdapter) CallAPI(endpoint string, params map[string]interface{}) (map[string]interface{}, error) {
	// 创建调用脚本
	scriptContent := fmt.Sprintf(`
package main

import (
	"encoding/json"
	"fmt"
	"os"
	BsPaySdk "github.com/huifurepo/bspay-go-sdk"
)

func main() {
	// 初始化SDK
	sdk, err := BsPaySdk.NewBsPay(%v, "%s")
	if err != nil {
		fmt.Fprintf(os.Stderr, "SDK init error: %%v", err)
		os.Exit(1)
	}
	
	// 调用API
	// 这里需要根据实际的SDK方法进行调整
	_ = sdk
	
	// 返回模拟结果
	result := map[string]interface{}{
		"success": true,
		"message": "API call simulated",
		"endpoint": "%s",
	}
	
	jsonData, _ := json.Marshal(result)
	fmt.Print(string(jsonData))
}
`, a.isProduction, a.configPath, endpoint)
	
	// 创建临时Go文件
	tempFile := filepath.Join(os.TempDir(), "huifu_call.go")
	if err := os.WriteFile(tempFile, []byte(scriptContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write script: %v", err)
	}
	defer os.Remove(tempFile)
	
	// 执行脚本
	cmd := exec.Command("go", "run", tempFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("script execution failed: %v, output: %s", err, output)
	}
	
	// 解析结果
	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse result: %v", err)
	}
	
	return result, nil
}

// Cleanup 清理临时文件
func (a *BsPayAdapter) Cleanup() {
	if a.configPath != "" {
		os.Remove(a.configPath)
	}
}