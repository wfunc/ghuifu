// API基础URL
const API_BASE_URL = '/api';

// 工具函数：显示提示信息
function showAlert(message, type = 'info') {
    const alertBox = document.getElementById('alertBox');
    alertBox.className = `alert alert-${type}`;
    // 使用新的CSS结构确保内容不会超出边界
    alertBox.innerHTML = `
        <div class="alert-content">
            <div class="alert-message">${message.replace(/\n/g, '<br>')}</div>
            <button class="close-btn" onclick="hideAlert()" title="点击关闭">×</button>
        </div>
    `;
    alertBox.style.display = 'block';
    alertBox.style.cursor = 'pointer';
    
    // 点击alert区域也可以关闭
    alertBox.onclick = hideAlert;
}

// 隐藏提示信息
function hideAlert() {
    const alertBox = document.getElementById('alertBox');
    alertBox.style.display = 'none';
}

// 工具函数：显示加载状态
function showLoading(show = true) {
    const loading = document.getElementById('loading');
    loading.style.display = show ? 'block' : 'none';
}

// 清空配置表单
function clearForm() {
    document.getElementById('configForm').reset();
    showAlert('表单已清空', 'info');
}

// 生成测试RSA密钥
async function generateTestKey() {
    try {
        // 调用后端API生成测试密钥
        const response = await fetch(`${API_BASE_URL}/generate-test-key`);
        if (response.ok) {
            const data = await response.json();
            document.getElementById('rsa_private_key').value = data.private_key;
            showAlert('已生成测试密钥（仅供测试，请勿用于生产环境）', 'info');
        } else {
            // 使用备用的静态测试密钥
            const testKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAvh5iez7A2Vt9f7B+qXc2O3z6Hn+fXsRp39SE+AmC7MmObRga
wLFXv74yMSDMIirHt26IOyM8Vu3HGIezI7tQqcYli1HJcyGovGYk4Lc5z/EB+Umg
sjA/lMrBdjN3EMjZ1xpVwu9DHCFey5cOCe2EqCMyeQ28Yy3mcACKqX2YDbrRN9re
o23uJtMwV1NsN/XSlEFY1NfbS/YWpZdJCHUo/f5asP9Rs+KTNGGUDrIQjUrHRq72
irw6kPnSCR34zuHNRpfEUrvHnPSuIcOZoSao1rAh1t4EooiBFeRk0cFbxrztxGTb
OFPDQOYCNd/JFPnYgS8QBPLcteUHzCytCL7ThwIDAQABAoIBABSNml3yicy1vFqK
jRbrAVzrBOs5JtSK7Vs6UWmzNYk9vP0ERxgf0/mxqSFwh0EGWPL2qxmhlItdR1Ha
kb5CKNVBy5tFKz8cG27KqB/3DvPw/SKjGBFLcAQ46zNJGw0geZRsj2r0jM/Et8fQ
u77NA7Ndor49gulB9BCVrmfmYQLSyMdmZDNp/VYGNAOX3dowrbUA4iLXj0Kvv6bY
-----END RSA PRIVATE KEY-----`;

            document.getElementById('rsa_private_key').value = testKey;
            showAlert('已生成测试密钥（仅供测试，请勿用于生产环境）', 'info');
        }
    } catch (error) {
        console.error('生成测试密钥失败:', error);
        showAlert('生成失败，请手动输入密钥', 'error');
    }
}

// 加载配置列表
async function loadConfigs() {
    try {
        showLoading(true);
        const response = await fetch(`${API_BASE_URL}/configs`);
        const data = await response.json();

        const configList = document.getElementById('configList');
        const sysIdSelect = document.getElementById('wx_sys_id');

        // 清空现有内容
        configList.innerHTML = '';
        sysIdSelect.innerHTML = '<option value="">请选择系统配置</option>';

        if (data.configs && data.configs.length > 0) {
            data.configs.forEach(config => {
                // 添加到配置列表
                const configItem = document.createElement('div');
                configItem.className = 'config-item';
                configItem.innerHTML = `
                    <div class="config-info">
                        <strong>系统ID:</strong> ${config.sys_id}
                        <span class="tag tag-${config.environment === 'production' ? 'prod' : 'test'}">
                            ${config.environment === 'production' ? '生产' : '测试'}环境
                        </span>
                        <br>
                        <small>产品ID: ${config.product_id}</small>
                    </div>
                    <div class="config-actions">
                        <button class="btn-secondary" onclick="selectConfig('${config.sys_id}')">选择</button>
                        <button class="btn-danger" onclick="deleteConfig('${config.sys_id}')">删除</button>
                    </div>
                `;
                configList.appendChild(configItem);

                // 添加到下拉选择
                const option = document.createElement('option');
                option.value = config.sys_id;
                option.textContent = `${config.sys_id} (${config.environment})`;
                sysIdSelect.appendChild(option);
            });
        } else {
            configList.innerHTML = '<p style="text-align: center; color: #999;">暂无配置</p>';
        }
    } catch (error) {
        console.error('加载配置失败:', error);
        showAlert('加载配置列表失败', 'error');
    } finally {
        showLoading(false);
    }
}

// 选择配置
function selectConfig(sysId) {
    document.getElementById('wx_sys_id').value = sysId;
    showAlert(`已选择配置: ${sysId}`, 'success');

    // 滚动到微信配置区域
    document.querySelector('#wechatForm').scrollIntoView({ behavior: 'smooth' });
}

// 删除配置
async function deleteConfig(sysId) {
    // 确认删除
    if (!confirm(`确定要删除配置 ${sysId} 吗？此操作不可恢复。`)) {
        return;
    }

    try {
        showLoading(true);
        const response = await fetch(`${API_BASE_URL}/config/${encodeURIComponent(sysId)}`, {
            method: 'DELETE',
            headers: {
                'Content-Type': 'application/json',
            }
        });

        const data = await response.json();

        if (response.ok) {
            showAlert(`配置 ${sysId} 已删除`, 'success');
            // 重新加载配置列表
            await loadConfigs();

            // 如果删除的是当前选中的配置，清空选择
            if (document.getElementById('wx_sys_id').value === sysId) {
                document.getElementById('wx_sys_id').value = '';
            }
        } else {
            showAlert(`删除失败: ${data.details || '未知错误'}`, 'error');
        }
    } catch (error) {
        console.error('删除配置失败:', error);
        showAlert('网络错误，请稍后重试', 'error');
    } finally {
        showLoading(false);
    }
}

// 保存配置
document.getElementById('configForm').addEventListener('submit', async (e) => {
    e.preventDefault();

    const formData = new FormData(e.target);
    const config = {
        sys_id: formData.get('sys_id'),
        product_id: formData.get('product_id'),
        rsa_private_key: formData.get('rsa_private_key'),
        environment: formData.get('environment'),
        // 这两个字段暂时留空，后续从微信配置表单获取
        wx_woa_app_id: '',
        wx_woa_path: ''
    };

    try {
        showLoading(true);
        const response = await fetch(`${API_BASE_URL}/config`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(config)
        });

        const data = await response.json();

        if (response.ok) {
            showAlert('配置保存成功！', 'success');
            // 清空表单
            document.getElementById('configForm').reset();
            // 重新加载配置列表
            await loadConfigs();
            // 自动选择刚保存的配置
            setTimeout(() => {
                document.getElementById('wx_sys_id').value = config.sys_id;
            }, 100);
        } else {
            showAlert(`保存失败: ${data.details || '未知错误'}`, 'error');
            console.error('保存失败详情:', data);
        }
    } catch (error) {
        console.error('保存配置失败:', error);
        showAlert('网络错误，请稍后重试', 'error');
    } finally {
        showLoading(false);
    }
});

// 配置微信商户
document.getElementById('wechatForm').addEventListener('submit', async (e) => {
    e.preventDefault();

    const formData = new FormData(e.target);
    const wechatConfig = {
        sys_id: formData.get('wx_sys_id'),
        huifu_id: formData.get('huifu_id'),
        wx_woa_app_id: formData.get('wx_woa_app_id'),
        wx_woa_path: formData.get('wx_woa_path'),
        fee_type: formData.get('fee_type')
    };

    if (!wechatConfig.sys_id) {
        showAlert('请先选择系统配置', 'error');
        return;
    }

    if (!wechatConfig.fee_type) {
        showAlert('请选择费率类型', 'error');
        return;
    }

    try {
        showLoading(true);
        const response = await fetch(`${API_BASE_URL}/wechat-config`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(wechatConfig)
        });

        const data = await response.json();

        if (response.ok) {
            // 显示配置结果
            let resultMessage = '✅ 微信商户配置成功！\n';
            resultMessage += `汇付ID: ${data.huifu_id}\n`;
            resultMessage += `微信AppID: ${data.wx_app_id}`;

            // 如果有额外的响应信息，显示它
            if (data.message) {
                // 如果message是对象，转换为字符串
                if (typeof data.message === 'object') {
                    const msgData = data.message.data || data.message;
                    if (msgData && msgData.resp_desc) {
                        resultMessage += `\n状态: ${msgData.resp_desc}`;
                    }
                } else {
                    resultMessage += `\n响应: ${data.message}`;
                }
            }

            showAlert(resultMessage, 'success');
        } else {
            showAlert(`配置失败: ${data.details || '未知错误'}`, 'error');
        }
    } catch (error) {
        console.error('配置微信商户失败:', error);
        showAlert('网络错误，请稍后重试', 'error');
    } finally {
        showLoading(false);
    }
});

// 查询微信商户配置
async function queryWeChatConfig() {
    const sysId = document.getElementById('wx_sys_id').value;
    const huifuId = document.getElementById('huifu_id').value;

    if (!sysId) {
        showAlert('请先选择系统配置', 'error');
        return;
    }

    if (!huifuId) {
        showAlert('请输入汇付ID', 'error');
        return;
    }

    try {
        showLoading(true);
        const response = await fetch(`${API_BASE_URL}/wechat-config-query`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                sys_id: sysId,
                huifu_id: huifuId
            })
        });

        const data = await response.json();

        if (response.ok) {
            // 格式化查询结果
            let resultMessage = '🔍 查询结果：\n';
            resultMessage += `汇付ID: ${data.huifu_id}\n`;

            if (data.message) {
                resultMessage += `响应: ${data.message}`;
            } else if (data.result) {
                // 如果result不是预期格式，显示原始数据
                resultMessage += `响应: ${JSON.stringify(data.result, null, 2)}`;
            }

            showAlert(resultMessage, 'success');
        } else {
            showAlert(`查询失败: ${data.details || '未知错误'}`, 'error');
        }
    } catch (error) {
        console.error('查询配置失败:', error);
        showAlert('查询失败，请检查网络连接', 'error');
    } finally {
        showLoading(false);
    }
}

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', () => {
    loadConfigs();

    // 自动刷新配置列表（每30秒）
    setInterval(loadConfigs, 30000);

    // 从URL参数自动填充表单
    const urlParams = new URLSearchParams(window.location.search);
    if (urlParams.has('sys_id')) {
        document.getElementById('sys_id').value = urlParams.get('sys_id');
    }
    if (urlParams.has('product_id')) {
        document.getElementById('product_id').value = urlParams.get('product_id');
    }
    if (urlParams.has('rsa_private_key')) {
        document.getElementById('rsa_private_key').value = urlParams.get('rsa_private_key');
    }
    if (urlParams.has('environment')) {
        document.getElementById('environment').value = urlParams.get('environment');
    }

    // RSA私钥输入框自动格式化
    const rsaInput = document.getElementById('rsa_private_key');
    rsaInput.addEventListener('paste', (e) => {
        setTimeout(() => {
            // 确保私钥格式正确
            let value = rsaInput.value;
            if (!value.includes('BEGIN') && !value.includes('END')) {
                value = `-----BEGIN RSA PRIVATE KEY-----\n${value}\n-----END RSA PRIVATE KEY-----`;
                rsaInput.value = value;
            }
        }, 10);
    });
});

// 添加键盘快捷键
document.addEventListener('keydown', (e) => {
    // Ctrl+S 保存配置
    if (e.ctrlKey && e.key === 's') {
        e.preventDefault();
        document.getElementById('configForm').dispatchEvent(new Event('submit'));
    }

    // Ctrl+R 刷新配置列表
    if (e.ctrlKey && e.key === 'r') {
        e.preventDefault();
        loadConfigs();
    }
});