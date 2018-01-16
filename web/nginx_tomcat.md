# 使用nginx和tomcat做负载均衡

## 背景

有哥们要搭个tomcat集群来进行数据并发测试，其中遇到了一些问题。

## 实践

* 启动两个tomcat
* 配置nginx
* 通过web浏览器进行验证


nginx的配置文件：
```
worker_processes  1;

events {
    worker_connections  1024;
}

http {
    upstream bogon {
        server 192.168.176.17:8080;
        server 192.168.176.20:8082;
    }

    server {
        listen       8188;
        server_name  bogon1;

        location / {
            proxy_pass http://bogon;
            index  index.html index.htm;
        }

        error_page   500 502 503 504  /50x.html;
        location = /50x.html {
            root   html;
        }

      }
}
```
## 遇到的问题

### connection refused while upstream

定位： nginx连接tomcat有问题。

经核对发现是因为http.server.location.proxy_pass 没有跟 http.upstream后面定义的名字对应，所以nginx找不到对应的tomcat

## 参考资料
[使用Nginx实现Tomcat集群负载均衡](http://www.cnblogs.com/machanghai/p/5957086.html)
