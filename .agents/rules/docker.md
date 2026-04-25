---
trigger: always_on
---

始终使用github-cli 来管理代码
每次Dockerfile有修改就出发github workflow build image
每个目录都是一个独立的docker 项目，配备独立的readme.md文件，镜像自动生成 在 ghcr.io/yorkane/xxxx 除非代码有tag, 否则始终是latest

