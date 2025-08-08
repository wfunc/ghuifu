#!/bin/bash

# 测试汇付配置系统的完整流程

echo "========================================="
echo "  汇付配置系统测试"
echo "========================================="

# 1. 生成测试密钥
echo -e "\n1. 生成测试密钥..."
KEY_RESPONSE=$(curl -s http://localhost:8080/api/generate-test-key)
PRIVATE_KEY=$(echo $KEY_RESPONSE | jq -r '.private_key')
echo "✅ 测试密钥生成成功"

# 2. 保存配置
echo -e "\n2. 保存配置..."
CONFIG_DATA=$(jq -n \
  --arg sys_id "test_system_001" \
  --arg product_id "test_product_001" \
  --arg rsa_private_key "$PRIVATE_KEY" \
  --arg wx_woa_app_id "wx12345678" \
  --arg wx_woa_path "pages/index/index" \
  --arg environment "test" \
  '{sys_id: $sys_id, product_id: $product_id, rsa_private_key: $rsa_private_key, wx_woa_app_id: $wx_woa_app_id, wx_woa_path: $wx_woa_path, environment: $environment}')

SAVE_RESPONSE=$(curl -s -X POST http://localhost:8080/api/config \
  -H "Content-Type: application/json" \
  -d "$CONFIG_DATA")

if echo "$SAVE_RESPONSE" | grep -q "success"; then
    echo "✅ 配置保存成功"
    echo "$SAVE_RESPONSE" | jq
else
    echo "❌ 配置保存失败"
    echo "$SAVE_RESPONSE" | jq
    exit 1
fi

# 3. 获取配置列表
echo -e "\n3. 获取配置列表..."
CONFIGS=$(curl -s http://localhost:8080/api/configs)
echo "$CONFIGS" | jq

# 4. 测试配置
echo -e "\n4. 测试配置..."
TEST_RESPONSE=$(curl -s -X POST http://localhost:8080/api/test-config \
  -H "Content-Type: application/json" \
  -d '{"sys_id": "test_system_001"}')

if echo "$TEST_RESPONSE" | grep -q "valid"; then
    echo "✅ 配置测试成功"
    echo "$TEST_RESPONSE" | jq
else
    echo "❌ 配置测试失败"
    echo "$TEST_RESPONSE" | jq
fi

# 5. 配置微信商户
echo -e "\n5. 配置微信商户..."
WECHAT_CONFIG=$(cat <<EOF
{
  "sys_id": "test_system_001",
  "huifu_id": "6666000108854952",
  "wx_woa_app_id": "wx98765432",
  "wx_woa_path": "pages/pay/index"
}
EOF
)

WECHAT_RESPONSE=$(curl -s -X POST http://localhost:8080/api/wechat-config \
  -H "Content-Type: application/json" \
  -d "$WECHAT_CONFIG")

echo "$WECHAT_RESPONSE" | jq

echo -e "\n========================================="
echo "  测试完成！"
echo "========================================="
echo -e "\n访问 http://localhost:8080 查看Web界面"