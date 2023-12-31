## 这里记录一些插件的指令

**🔔 注意：以下`[]`符号内代表此处是需要填写内容，并且不需要`[]`符号**

### 内置指令

**🚀 这部分指令都只能由管理员使用，并且都有前缀，前缀在`config.yaml`中设置`commandPrefix`**

* `/响应 [个人wxId | 群wxId | (留空代表当前聊天室)]`
    * 该指令用于设置机器人在某个聊天室内开始响应
* `/沉默 [个人wxId | 群wxId | (留空代表当前聊天室)]`
    * 该指令用于设置机器人在某个聊天室内开始沉默，即不处理任何插件指令
* `/启用 [插件服务名(英文那个，发送菜单可以看到)]`
    * 该指令用于启用某个插件服务，默认在当前会话聊天室
* `/禁用 [插件服务名(英文那个，发送菜单可以看到)]`
    * 该指令用于禁用某个插件服务，默认在当前会话聊天室
* `/启用全部 [插件服务名]`
    * 该指令用于启用某个插件服务，对所有聊天室生效
* `/禁用全部 [插件服务名]`
    * 该指令用于禁用某个插件服务，对所有聊天室生效
* `/ban [插件服务名] [用户微信ID]` 或 在群里引用对方消息回复`/ban [插件服务名]`即可快速ban掉
  * 该指令用于ban掉某个用户在某个群里发指令响应
* `/unban [插件服务名] [用户微信ID]` 或 在群里引用对方消息回复`/unban [插件服务名]`即可快速unban掉
  * 该指令用于unban掉某个用户在某个群里发指令响应
### 插件指令

[参考各插件目录下`README.md`文件](../plugins)