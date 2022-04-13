## 说明

covid-tracker的作用是通过chrome-dep(可能是多此一举)爬取上海市卫健委疫情发布的消息，给所有订阅消息的用户发送自己所在区域是否有官方通报的新增阳性患者，以及上海市各辖区的疫情情况。

可以访问[上海新冠疫情订阅](http://dboat.cn/register)来订阅消息。订阅只需要提供住址以及邮箱地址即可。


## 如果你自己HOST

### Mac
0. 安装Chrome
1. 创建一个exmail.conf文件
2. ./covid-tracker -c {path to exmail.conf}

exmail.conf 是你用来发送邮件的配置文件，dboat.cn使用的是企业微信邮箱，因此配置文件格式是这样的:
```
{
    "Host":"imap.exmail.qq.com:993",
    "TLS":true,
    "InsecureSkipVerify":true,
    "User":"",
    "Pwd":"",
    "Folder":"Inbox",
    "ReadOnly":true,
    "Username":""
}
```

### Linux

0. linux需要启动docker容器来支持chromdep

```
docker pull chromedp/headless-shell:latest
docker run -d -p 9222:9222 --rm --name headless-shell chromedp/headless-shell
```
剩余两步同上。
