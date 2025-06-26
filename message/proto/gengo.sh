set -e

# old
# # 采用docker proto镜像生成go代码
# docker run --rm -v `pwd`:/defs safetyculture/protoc-go
# 
# mv ./pb-go/*.go ./
# # 删除pb-go目录
# rm -r ./pb-go

# new
# 采用docker proto镜像生成go代码
 docker run --rm -v $(pwd):/defs rvolosatovs/protoc --proto_path=/defs --go_out=/defs /defs/*.proto

# # 采用docker proto镜像生成c++代码
#  docker run --rm -v $(pwd):/defs rvolosatovs/protoc --proto_path=/defs --cpp_out=/defs  /defs/*.proto

# # 采用docker proto镜像生成java代码
#  docker run --rm -v $(pwd):/defs rvolosatovs/protoc --proto_path=/defs --java_out=/defs  /defs/*.proto

# # 采用docker proto镜像生成rust代码
#  docker run --rm -v $(pwd):/defs rvolosatovs/protoc --proto_path=/defs --rust_out=/defs  /defs/*.proto

# 移动代码到当前目录
 mv ./msgproto/*.go ./
 # 删除L临时目录
 rm -r ./msgproto