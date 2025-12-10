# 需求说明

`cloudctl aws cdn` 新增一条子命令，根据配置文件批量创建 cdn，要求如下：

1. 配置文件采用 yaml 格式
2. 自定义的内容：
   2.1 常规配置：

- name: 这里用域名，比如 `test1.com`
- 备用域名：这里可以配置多个域名，比如 `test1.com`，`www.test1.com`，应该是一个列表
- 自定义证书：这里应该是使用证书 id?请帮我确定

  2.2 安全性：

- waf: 这里需要指定 waf id?请帮我确定

  2.3 源：

- 源的值：比如 alb 域名 `prod-web-749459849.ap-east-1.elb.amazonaws.com` ,是否还需要其它值

  2.4 行为：这里添加两条行为：

- 优先级 0，path:/app-api/\* , 将 HTTP 重定向到 HTTPS, Managed-CachingDisabled, Managed-AllViewer,Managed-SimpleCORS
- 优先级 1，默认（\*） , 将 HTTP 重定向到 HTTPS, Managed-CachingOptimized, Managed-AllViewer,Managed-SimpleCORS

其余保持不变
