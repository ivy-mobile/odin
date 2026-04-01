set -e

# 采用docker proto镜像生成go代码
 docker run --rm -v $(pwd):/defs rvolosatovs/protoc --proto_path=/defs --go_out=/defs /defs/*.proto

 # 移动代码到当前目录
  mv ./envelope/*.go ./
  # 删除L临时目录
  rm -r ./envelope