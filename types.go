package main

// ConfigRequest 配置请求结构体
type ConfigRequest struct {
	SysID         string `json:"sys_id" binding:"required"`
	ProductID     string `json:"product_id" binding:"required"`
	RSAPrivateKey string `json:"rsa_private_key" binding:"required"`
	WxWoaAppID    string `json:"wx_woa_app_id"`    // 可选，微信小程序AppID
	WxWoaPath     string `json:"wx_woa_path"`     // 可选，微信小程序路径
	Environment   string `json:"environment"` // production or test
}

// HuifuClient SDK客户端接口
type HuifuClient interface {
	CallAPI(endpoint string, params map[string]interface{}) (map[string]interface{}, error)
}