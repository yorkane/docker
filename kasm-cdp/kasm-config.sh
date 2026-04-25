#!/bin/bash
# KasmVNC 动态配置脚本
# 根据环境变量配置认证、帧率和画质

VNC_STARTUP="/dockerstartup/vnc_startup.sh"
USER_YAML="$HOME/.vnc/kasmvnc.yaml"

# ========== 1. VNC 认证配置 ==========
if [ -z "$VNC_PW" ]; then
    # 无密码 → 禁用认证
    sed -i 's/__KASM_AUTH_FLAGS__/-SecurityTypes None -DisableBasicAuth 1/g' "$VNC_STARTUP"
    echo "🔓 VNC 认证: 已禁用（VNC_PW 未设置）"
else
    # 有密码 → 启用认证（移除占位符，使用 Kasm 默认认证机制）
    sed -i 's/__KASM_AUTH_FLAGS__//g' "$VNC_STARTUP"
    echo "🔒 VNC 认证: 已启用"
fi

# ========== 2. VNC 画质配置 ==========
MAX_FPS="${KASM_MAX_FRAME_RATE:-30}"
JPEG_QUALITY="${KASM_JPEG_QUALITY:-7}"
MAX_QUALITY="${KASM_MAX_QUALITY:-8}"

echo "🎬 VNC 画质: ${MAX_FPS}fps | JPEG ${JPEG_QUALITY}/9 | 编码 ${MAX_QUALITY}/9"

# 创建用户级配置覆盖（优先级高于系统配置）
mkdir -p "$(dirname "$USER_YAML")"
cat > "$USER_YAML" <<EOF
encoding:
  max_frame_rate: ${MAX_FPS}
  rect_encoding_mode:
    min_quality: ${JPEG_QUALITY}
    max_quality: ${MAX_QUALITY}
    consider_lossless_quality: 10
  video_encoding_mode:
    jpeg_quality: ${JPEG_QUALITY}

runtime_configuration:
  allow_client_to_override_kasm_server_settings: false
EOF

echo "✅ KasmVNC 配置完成"
