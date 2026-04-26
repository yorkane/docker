#!/bin/bash
# Chromium Wrapper：自动重启循环 + CDP + 内存优化
# 适用于 2GB 内存服务器

CHROME_BIN=$(which chromium-browser 2>/dev/null || which chromium 2>/dev/null)

# 构建扩展加载参数（如有 CHROME_EXTENSIONS 环境变量或 /config/extensions 目录）
EXT_ARGS=""
if [ -n "$CHROME_EXTENSIONS" ]; then
    EXT_ARGS="--load-extension=$CHROME_EXTENSIONS"
elif [ -d "/config/extensions" ] && [ "$(ls -A /config/extensions 2>/dev/null)" ]; then
    EXT_DIRS=$(find /config/extensions -mindepth 1 -maxdepth 1 -type d | paste -sd, -)
    if [ -n "$EXT_DIRS" ]; then
        EXT_ARGS="--load-extension=$EXT_DIRS"
    fi
fi

if [ -z "$CHROME_BIN" ]; then
    echo "❌ Chromium 未找到"
    exit 1
fi

# GPU 参数：默认禁用 GPU，设置 CHROME_ENABLE_GPU=true 启用
GPU_ARGS=""
if [ "${CHROME_ENABLE_GPU}" = "true" ] || [ "${CHROME_ENABLE_GPU}" = "1" ]; then
    GPU_ARGS="--use-gl=angle --use-angle=swiftshader"
    echo "🎮 GPU 模式: SwiftShader (软件渲染)"
else
    GPU_ARGS="--disable-gpu --disable-gpu-compositing --disable-vulkan --disable-software-rasterizer --in-process-gpu --use-gl=disabled"
    echo "🖥️ GPU 模式: 已禁用 (纯 CPU)"
fi

while true; do
    echo "🚀 启动 Chromium (CDP :19222)..."
    $CHROME_BIN \
        --no-sandbox \
        $GPU_ARGS \
        --remote-debugging-port=19222 \
        --remote-debugging-address=127.0.0.1 \
        --remote-allow-origins="*" \
        --user-data-dir=/config/chromium-data \
        --disable-dev-shm-usage \
        --disable-background-networking \
        --disable-default-apps \
        --disable-sync \
        --disable-translate \
        --no-first-run \
        --no-default-browser-check \
        --disable-backgrounding-occluded-windows \
        --renderer-process-limit=2 \
        --js-flags="--max-old-space-size=512" \
        --disk-cache-size=33554432 \
        --disable-breakpad \
        --lang=zh-CN \
        $EXT_ARGS \
        about:blank 2>&1
    EXIT_CODE=$?
    echo "⚠️ Chromium 已退出 (code=$EXIT_CODE)，2 秒后自动重启..."
    sleep 2
done
