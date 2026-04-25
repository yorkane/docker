#!/bin/bash
# KasmVNC 动态配置脚本
# 根据环境变量覆盖 kasmvnc.yaml 中的帧率和画质设置

YAML_FILE="/etc/kasmvnc/kasmvnc.yaml"
USER_YAML="$HOME/.vnc/kasmvnc.yaml"

MAX_FPS="${KASM_MAX_FRAME_RATE:-30}"
JPEG_QUALITY="${KASM_JPEG_QUALITY:-7}"
MAX_QUALITY="${KASM_MAX_QUALITY:-8}"

echo "🎬 KasmVNC 配置:"
echo "   帧率: ${MAX_FPS} fps"
echo "   JPEG 质量: ${JPEG_QUALITY}/9"
echo "   最大编码质量: ${MAX_QUALITY}/9"

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

echo "✅ KasmVNC 配置已写入 ${USER_YAML}"
