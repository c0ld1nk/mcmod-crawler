# mcmod-crawler
用于爬取mcmod整合包数据，单文件自带web-ui，支持多种排序方式

爬取后会在本目录下生成`modpacks.json`，若存在此文件，不会重复爬取

每周固定爬取一次，暂时没有更改频率的功能

Example：
```shell
./mcmod-crawler
```

参数
```text
-thread 线程数，默认25，太高了可能会对服务器造成影响，适度
```

有什么需求和bug请在issue反馈