#!/bin/bash
set -e

# 定义目标目录
TARGET_DIR="mystate-chart/templates"

# 遍历目录下的所有yaml文件
for file in "$TARGET_DIR"/*.yaml; do
  # 检查文件是否存在
  if [[ -f "$file" ]]; then
    # 使用sed命令替换fullname为name
    sed -i 's/{{ include "mystate-chart.fullname" . }}/{{ .Release.Name }}/g' "$file"
    echo "Processed $file"
  else
    echo "No YAML files found in $TARGET_DIR"
  fi
done

repository=dayeguilaiye/my-state

sudo docker build -t $repository:latest .
sudo docker push $repository:latest

KUBECONFIG=/etc/rancher/k3s/k3s.yaml /usr/local/bin/helm upgrade --install my-state-release ./mystate-chart --set controllerManager.manager.image.repository=$repository --set controllerManager.manager.imagePullPolicy=Never