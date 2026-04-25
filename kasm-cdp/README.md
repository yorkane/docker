# Kasm Chrome with CDP (kasm-cdp)

This Docker image extends the official `kasmweb/chrome:develop` image by exposing the Chrome DevTools Protocol (CDP) port. It allows you to run a headless Chrome browser with a VNC interface for debugging, while simultaneously controlling it via automation tools like Puppeteer, Playwright, or Selenium using CDP.

## Features

- **Base Image**: `kasmweb/chrome:develop` (Ubuntu-based)
- **CDP Exposed**: Chrome DevTools Protocol is exposed on port `9222`.
- **VNC Access**: Accessible via a web browser on port `6901`.
- **Customizable Performance**: Dynamic configuration of VNC framerate and image quality via environment variables.
- **Security Adjustments**: SSL and basic authentication for VNC are disabled by default for easier local access and debugging.

## Environment Variables

You can customize the KasmVNC performance by setting the following environment variables when running the container:

| Variable | Default | Description |
|---|---|---|
| `KASM_MAX_FRAME_RATE` | `30` | Maximum VNC framerate. |
| `KASM_JPEG_QUALITY` | `7` | JPEG compression quality (0-9, higher is better). |
| `KASM_MAX_QUALITY` | `8` | Maximum encoding quality (0-9). |

## Usage

### 1. Build the Image

*(Optional)* If you want to build the image locally instead of pulling from the registry:

```bash
docker build -t ghcr.io/yorkane/kasm-cdp:latest kasm-cdp/
```

### 2. Push the Image

*(Optional)* To push the image to the GitHub Container Registry:

```bash
docker push ghcr.io/yorkane/kasm-cdp:latest
```
*(Note: A GitHub Actions workflow is configured to automatically build and push this image upon commits to the `kasm-cdp/` directory.)*

### 3. Run the Container

Run the container, mapping the necessary ports (6901 for VNC, 9222 for CDP). The `--shm-size` parameter is highly recommended to prevent Chrome from crashing.

```bash
docker run -d \
  --name kasm-cdp \
  --shm-size=512m \
  -p 6901:6901 \
  -p 9222:9222 \
  -e VNC_PW=password \
  ghcr.io/yorkane/kasm-cdp:latest
```

**With customized performance settings:**

```bash
docker run -d \
  --name kasm-cdp \
  --shm-size=512m \
  -p 6901:6901 \
  -p 9222:9222 \
  -e VNC_PW=password \
  -e KASM_MAX_FRAME_RATE=60 \
  -e KASM_JPEG_QUALITY=9 \
  -e KASM_MAX_QUALITY=9 \
  ghcr.io/yorkane/kasm-cdp:latest
```

## Accessing the Services

- **VNC Web Interface**: Open a browser and navigate to `http://localhost:6901`. (No password required due to disabled auth).
- **CDP JSON Endpoint**: Test CDP availability by visiting `http://localhost:9222/json/version`.
- **CDP WebSocket**: Connect your automation tools to the WebSocket URL provided by the JSON endpoint or use the generic endpoint like `ws://localhost:9222/devtools/browser/<id>`.
