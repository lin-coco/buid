# buid

一个轻量级高性能带有缓冲的全局唯一id生成器


分布式唯一id要求满足：
1. 唯一性：生成id全局唯一
2. 单调递增：生成id后一个比前一个要大，便于数据库插入
3. 安全性：不暴露用户数和订单数等信息，id最好不能连续

相比于snowflake优点：
1. 性能优
2. 解决时钟回拨
3. 最长连续id不超过2
4. 缓冲区应对突发流量
5. 可配置

缺点： 无机器id（但可作为服务提供给其他服务使用）


1. 作为库使用
```go
func main() {
   buid := internal.NewDefaultBUID()
   buid.StartGenerate()

   now := time.Now()
   _ = buid.Gets(1000000)
   fmt.Println(time.Since(now).Milliseconds())
   buid.Exit()
}
```
   
2. 作为服务使用
```go
//...
```


