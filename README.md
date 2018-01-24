# XRtspServer


## Rtsp服务器

#示例代码见

https://github.com/NiuStar/RtspServerTest
#

## 目前已实现功能：

1、RTSP接收推流

2、RTSP直播

3、API接口

通过调用RtspClientManager.GetCurrManagers()获取目前观看者列表

```Json
[{

"ClientS": ["192.168.1.92:56181"],
"Url": "rtsp://192.168.1.92:8554/1_s.sdp"

}]
```

数组内为推送视频流列表，

URL：视频流的RTSP地址，

ClientS：观看者的源地址列表

## 下一步开发目标：


1、后台配置管理页面

2、视频流切片存储功能

3、HLS转发功能

4、RTMP推流功能

5、RTMP直播功能

6、MP4视频直播

## 欢迎有兴趣爱好者一起加盟开发

联系方式：24802117（QQ）

邮箱：24802117@qq.com



