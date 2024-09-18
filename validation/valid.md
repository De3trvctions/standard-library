# ValidV2 使用方法

![Validate Version](https://img.shields.io/badge/Version-v0.1-brightgreen?style=plastic) ![Markdown Version](https://img.shields.io/badge/MdVersion-1.0-blue?style=plastic)

## <span style="color:red">首要须知</span>

在 ValidV2 里，所有函数都是通过 Beego Validation 包裹衍生出来的。
可以调用 Beego Validation 自带的校验包或者通过 Beego Validation 里的 `AddCustomFunc` 来新增自定义的检验方式。

同时间，我们也可以覆盖重写 Beego Validation 里自带的校验包，但是此方法不被推荐。在不是没办法的情况下，还是不要覆盖重写原厂包的校验包。

---

## 详情内容

### 在哪触发

校验调用应该只会在 gRPC 的 server 部分。原因就是因为在服务里调用校验是一个相对来说奇怪的事情。因为要把数据传到服务里面再校验也没有问题再返还。
校验的目的就是为了减轻服务的工作量，当数据传入有问题的时候直接报错就会节省很多数据流。

### 如何触发

触发校验是通过每一个结构体`struct`里面的字符串`Tag`设置。
校验引擎会查找字符串里的`valid`字眼，然后调取里面所写的校验是什么校验。
例如：

```go
type Search struct {
    SearchWord              string `valid:"IsDescription"`
    Page                    int64  `valid:"Min(1)"`
    PageSize                int64  `valid:"Range(0,2000)"`
    SortingMethod           string `valid:"Alpha"`
    AdminName               string
}
```

通过设置字符串里面的校验方式，达成调取校验该参数。须谨慎的是，在填写校验的时候，需知道结构体是否是通用结构体，每一个参数是否会传不一样的值 (尤其是`string`的参数值)， 这是因为`string`是可以拥有各种组合模式的校验，也是较为复杂的校验。

---

**更多阅读**
[Beego Validation 华语版](https://www.kancloud.cn/hello123/beego/126131#API__130)
[Beego Validation 官网版](https://pkg.go.dev/github.com/astaxie/beego/validation)
[RegExp](https://pkg.go.dev/regexp/syntax@go1.20)
[Unicode for RegExp](https://pkg.go.dev/unicode)
