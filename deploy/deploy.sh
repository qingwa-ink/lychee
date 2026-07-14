#!/bin/bash
set -e
# lychee 部署到测试环境 (lychee.uvera.ai)
# 在 lychee 项目根目录运行此脚本
KEY=~/Downloads/uvera.pem
HOST=ubuntu@ec2-100-55-48-57.compute-1.amazonaws.com

echo "=== 1. 创建源代码 tarball（排除 config 和数据目录） ==="
tar --exclude='.git' --exclude='bin' --exclude='.DS_Store' --exclude='deploy/config' --exclude='deploy/data' -czf /tmp/lychee-src.tar.gz .

echo "=== 2. 上传 tarball 到 EC2 ==="
scp -i "$KEY" /tmp/lychee-src.tar.gz "$HOST":/home/ubuntu/lychee/

echo "=== 3. 在 EC2 上解压并构建 Docker 镜像 ==="
ssh -i "$KEY" "$HOST" "
cd /home/ubuntu/lychee
tar xzf lychee-src.tar.gz
docker build -f deploy/Dockerfile -t lychee:latest .
"

echo "=== 4. 添加 nginx 配置并 reload ==="
ssh -i "$KEY" "$HOST" "
cp /home/ubuntu/lychee/deploy/nginx-lychee.conf /home/ubuntu/uvera-test-release/deploy/nginx/conf.d/lychee.conf
docker exec uvera-test-gateway nginx -t
docker exec uvera-test-gateway nginx -s reload
"

echo "=== 5. 停止旧容器（如有）==="
ssh -i "$KEY" "$HOST" "
cd /home/ubuntu/lychee
docker compose -f deploy/docker-compose.yml down 2>/dev/null || true
"

echo "=== 6. 上传 config 和数据文件 ==="
ssh -i "$KEY" "$HOST" "mkdir -p /home/ubuntu/lychee/deploy/config /home/ubuntu/lychee/deploy/data"
scp -i "$KEY" config_test.yaml "$HOST":/home/ubuntu/lychee/deploy/config/config.yaml
scp -i "$KEY" data/lychee.db "$HOST":/home/ubuntu/lychee/deploy/data/lychee.db

echo "=== 7. 校验上传文件 ==="
ssh -i "$KEY" "$HOST" "
test -f /home/ubuntu/lychee/deploy/config/config.yaml || { echo 'ERROR: config.yaml missing!'; exit 1; }
test -f /home/ubuntu/lychee/deploy/data/lychee.db || { echo 'ERROR: lychee.db missing!'; exit 1; }
echo '  config.yaml: ' \$(stat -c%s /home/ubuntu/lychee/deploy/config/config.yaml) ' bytes'
echo '  lychee.db:   ' \$(stat -c%s /home/ubuntu/lychee/deploy/data/lychee.db) ' bytes'
"

echo "=== 8. 启动 lychee 服务 ==="
ssh -i "$KEY" "$HOST" "
cd /home/ubuntu/lychee
docker compose -f deploy/docker-compose.yml up -d
"

echo "=== 9. 等待服务启动 ==="
sleep 3

echo "=== 10. 验证 ==="
echo "--- 容器状态 ---"
ssh -i "$KEY" "$HOST" "docker ps --filter name=lychee-test --format '{{.Status}}'"
echo "--- 容器日志 ---"
ssh -i "$KEY" "$HOST" "docker logs lychee-test --tail 10"
echo "--- nginx 路由测试 ---"
ssh -i "$KEY" "$HOST" "curl -s -o /dev/null -w '%{http_code}' -H 'Host: lychee.uvera.ai' http://localhost/login"
echo ""
echo "=== 部署完成 ==="
echo "访问: https://lychee.uvera.ai/login"
