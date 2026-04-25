#!/usr/bin/env python3
"""
CDP Auth Proxy — 为 Chrome DevTools Protocol 提供 Token 认证

- 私有网段 (127.0.0.0/8, 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16) 免认证
- 外部 IP 必须通过 URL 参数 ?token=xxx 或 Authorization: Bearer xxx 认证
- 未设置 CDP_TOKEN 时，所有请求均放行（向后兼容）
"""

import asyncio
import ipaddress
import os
import sys
from urllib.parse import urlparse, parse_qs

TOKEN = os.environ.get("CDP_TOKEN", "")
UPSTREAM_HOST = "127.0.0.1"
UPSTREAM_PORT = 19222
LISTEN_PORT = 9222

PRIVATE_NETS = [
    ipaddress.ip_network("127.0.0.0/8"),
    ipaddress.ip_network("10.0.0.0/8"),
    ipaddress.ip_network("172.16.0.0/12"),
    ipaddress.ip_network("192.168.0.0/16"),
]


def is_private(ip_str: str) -> bool:
    """判断 IP 是否属于私有网段"""
    try:
        addr = ipaddress.ip_address(ip_str)
        return any(addr in net for net in PRIVATE_NETS)
    except ValueError:
        return False


def check_token(header_data: bytes) -> bool:
    """从 HTTP 请求头中提取并校验 token"""
    try:
        text = header_data.decode("utf-8", errors="replace")
        lines = text.split("\r\n")

        # 从 URL query 参数中获取 token
        # e.g. GET /json/version?token=xxx HTTP/1.1
        request_line = lines[0] if lines else ""
        parts = request_line.split(" ")
        if len(parts) >= 2:
            parsed = urlparse(parts[1])
            params = parse_qs(parsed.query)
            url_token = params.get("token", [None])[0]
            if url_token == TOKEN:
                return True

        # 从 Authorization header 中获取 token
        # e.g. Authorization: Bearer xxx
        for line in lines[1:]:
            if line.lower().startswith("authorization:"):
                val = line.split(":", 1)[1].strip()
                if val.lower().startswith("bearer "):
                    bearer_token = val[7:].strip()
                    if bearer_token == TOKEN:
                        return True
                break

        return False
    except Exception:
        return False


async def pipe(reader: asyncio.StreamReader, writer: asyncio.StreamWriter):
    """双向数据管道"""
    try:
        while True:
            data = await reader.read(65536)
            if not data:
                break
            writer.write(data)
            await writer.drain()
    except (ConnectionResetError, BrokenPipeError, ConnectionAbortedError):
        pass
    finally:
        try:
            writer.close()
        except Exception:
            pass


async def handle_client(client_reader: asyncio.StreamReader,
                        client_writer: asyncio.StreamWriter):
    """处理每个客户端连接"""
    peer = client_writer.get_extra_info("peername")
    client_ip = peer[0] if peer else "unknown"

    try:
        # 读取初始 HTTP 请求头
        header_data = await asyncio.wait_for(client_reader.read(8192), timeout=10)
        if not header_data:
            client_writer.close()
            return

        # 判断是否需要认证：设置了 TOKEN 且来源不是私有网段
        need_auth = bool(TOKEN) and not is_private(client_ip)

        if need_auth and not check_token(header_data):
            response = (
                b"HTTP/1.1 401 Unauthorized\r\n"
                b"Content-Type: text/plain; charset=utf-8\r\n"
                b"WWW-Authenticate: Bearer\r\n"
                b"Connection: close\r\n\r\n"
                b"401 Unauthorized: Invalid or missing CDP token\n"
            )
            client_writer.write(response)
            await client_writer.drain()
            client_writer.close()
            print(f"🚫 Rejected: {client_ip} (no valid token)")
            return

        # 连接上游 Chrome CDP
        upstream_reader, upstream_writer = await asyncio.open_connection(
            UPSTREAM_HOST, UPSTREAM_PORT
        )

        # 转发初始请求
        upstream_writer.write(header_data)
        await upstream_writer.drain()

        # 建立双向管道
        await asyncio.gather(
            pipe(client_reader, upstream_writer),
            pipe(upstream_reader, client_writer),
        )
    except asyncio.TimeoutError:
        print(f"⏱️  Timeout: {client_ip}")
    except ConnectionRefusedError:
        response = (
            b"HTTP/1.1 502 Bad Gateway\r\n"
            b"Content-Type: text/plain\r\n"
            b"Connection: close\r\n\r\n"
            b"502 Bad Gateway: Chrome CDP is not ready\n"
        )
        try:
            client_writer.write(response)
            await client_writer.drain()
        except Exception:
            pass
    except Exception as e:
        print(f"❌ Error handling {client_ip}: {e}")
    finally:
        try:
            client_writer.close()
        except Exception:
            pass


async def main():
    if not TOKEN:
        print("⚠️  CDP_TOKEN 未设置，所有访问均放行（无认证）")
    else:
        print(f"🔒 CDP Auth Proxy: Token 认证已启用，私有网段免认证")

    server = await asyncio.start_server(handle_client, "0.0.0.0", LISTEN_PORT)
    print(f"🚀 CDP Auth Proxy 监听 0.0.0.0:{LISTEN_PORT} -> "
          f"{UPSTREAM_HOST}:{UPSTREAM_PORT}")

    async with server:
        await server.serve_forever()


if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        sys.exit(0)
