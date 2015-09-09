# 构建镜像




### 功能

支持用git拉取最新代码，根据项目根目录Dockerfile文件创建镜像，并将镜像push到Registry中。






### 配置说明

- 仅支持container-topology的镜像构建（业务相关的代码才需频繁的镜像构建）；

- 若无repo配置字段（git地址），直接使用image的配置镜像；

- 若有repo配置字段，根据repo中的Dockerfile创建镜像，以image作为镜像名，git的commitID做为tag，push到Registry，image将被替换成新的镜像地址被后续使用；

