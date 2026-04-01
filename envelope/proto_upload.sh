#!/usr/bin/env bash

set -euo pipefail

# 配置项
TARGET_REPO="git@10.80.1.11:ivy-vs-backend/party-pop-idl.git"
TARGET_DIR="proto/game"
REPO_NAME="party-pop-idl"
COMMIT_MESSAGE="feat: update game envelope proto"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEMP_DIR=""

cleanup() {
  if [[ -n "${TEMP_DIR}" && -d "${TEMP_DIR}" ]]; then
    echo "[+] 清理临时目录..."
    rm -rf "${TEMP_DIR}"
  fi
}

trap cleanup EXIT

if ! command -v git >/dev/null 2>&1; then
  echo "错误: 未找到 git 命令，请先安装 Git 并确认其在 PATH 中。"
  exit 1
fi

echo "[+] 获取远程仓库分支列表..."
mapfile -t branches < <(git ls-remote --heads "${TARGET_REPO}" | awk '{sub("refs/heads/", "", $2); print $2}')

if [[ ${#branches[@]} -eq 0 ]]; then
  echo "错误: 未找到任何远程分支。"
  exit 1
fi

echo
echo "远程仓库分支列表："
for i in "${!branches[@]}"; do
  printf '%d. %s\n' "$((i + 1))" "${branches[$i]}"
done

echo
read -r -p "请选择分支序号 [1-${#branches[@]}]: " branch_choice

if ! [[ "${branch_choice}" =~ ^[0-9]+$ ]] || (( branch_choice < 1 || branch_choice > ${#branches[@]} )); then
  echo "错误: 无效的分支序号。"
  exit 1
fi

BRANCH="${branches[$((branch_choice - 1))]}"
echo "已选择分支: ${BRANCH}"

TEMP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/${REPO_NAME}.XXXXXX")"
CLONE_DIR="${TEMP_DIR}/${REPO_NAME}"

echo "[+] 克隆目标仓库..."
git clone --branch "${BRANCH}" --depth 1 "${TARGET_REPO}" "${CLONE_DIR}"

echo "[+] 确保目标目录存在..."
mkdir -p "${CLONE_DIR}/${TARGET_DIR}"

shopt -s nullglob
proto_files=("${SCRIPT_DIR}"/*.proto)
shopt -u nullglob

if [[ ${#proto_files[@]} -eq 0 ]]; then
  echo "错误: 未找到需要上传的 proto 文件。"
  exit 1
fi

echo "[+] 复制 proto 文件..."
cp "${proto_files[@]}" "${CLONE_DIR}/${TARGET_DIR}/"

cd "${CLONE_DIR}"
echo "[+] 检查更改..."
git status --short

git add "${TARGET_DIR}"

if git diff --cached --quiet; then
  echo "警告: 没有需要提交的更改，直接结束。"
  exit 0
fi

echo "[+] 提交更改..."
git commit -m "${COMMIT_MESSAGE}"

echo "[+] 推送更改..."
git push origin "${BRANCH}"

echo "操作完成！"
