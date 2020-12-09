## **Week 3**

### 1. 相关概念
#### 1.1 进程、线程与协程
- 进程: 
  - 是程序运行时的实例，也是操作系统资源分配的基本单元，拥有自己独立的地址空间
  - 生命周期从一个程序的启动到结束。
- 线程: 
  - 是CPU调度的最小单元
  - 负责执行进程的一次独立的计算任务，除寄存器和栈以外不占有独有资源，能访问进程拥有的资源
  - 线程的上下文切换需要陷入操作系统内核态
  - 生命周期往往较短，从进程创建线程开始，到线程执行完成退出
- 携程:
  - 轻量级线程，上下文切换等流程不需要陷入内核态，全在用户态运行，因此创建/销毁和切换等开销比线程更小

#### 1.2 并行与并发
- 并行: 同一时间可以有多个任务**同时**被运行
- 并发: 系统内可以同时存在多个执行中的任务
- 并行和并发的区别: 单核CPU能并发，但不能并行；多核CPU，既能并发，也能并行

### 2. Goroutine
Goroutine在类型上属于1.1中的携程
#### 2.1 创建并启动一个Goroutine
```
go doSomeThing()  // Explicit

// or 

go func() {
    // do something
}()               // Anonymous
```

#### 2.2 服务端Goroutine的一些案例和姿势
- 起Goroutine处理请求，但主线程用select{}阻塞不干事
  > Keep your self busy or do the work yourself
```
// 不推介，应该让主线程自己来处理
func main() {
    go func() {
        // serve http requests
    }()

    select {}
}
```
- 不要在接口耦合含Goroutine处理方式上的语义
  > Leave concurrency to the caller
```
func WalkDirectory(dir string) ([]string, error)
func WalkDirectory(dir string) chan string // 不推介, 强制约束了调用者必须读完chan的所有数据直到close，否则接口中的Goroutine将泄露
```
- 管理好Goroutine的生命周期
  > Never start a goroutine without knowning when it will stop
```
// 不推荐，servePortA或者servePortB任一方退出了，另外的一方无法感知，导致出现局部服务可用
func main() {
    go servePortA()
    go servePortB()
    select{}
}
```
```
// exit gracefully
func serve(addr string, stop <-chan struct{}>) error {
    go func() {
        <-stop
        // shutdown http server
    }()
    // handle http requests
}

func main() {
    var serverNum int = 2
    done := make(chan error, serverNum)
    stop := make(chan struct{})
    go func() {
        done <- serve("addrA", stop)
    }()
    go func() {
        done <- serve("addrB", stop)
    }()

    var stopped bool
    for i := 0; i < serverNum; i++ {
        if err <- done; err != nil {
        }
        if stopped != true {
            stopped = true
            close(stop)
        }
    }
}
```
- 使用sync.WaitGroup来协调多个Goroutine

#### 2.3 Goroutine泄露
Goroutine泄露体现为创建的Goroutine一直未退出且脱离了程序的控制范围，会造成Goroutine占用的内存资源得不到释放，当泄露得Goroutine达到一定数量时甚至导致内存耗尽程序崩溃。常见的Goroutine泄露的情况:
- Goroutine阻塞在Channel上
- 没有退出机制或出现死循环的Goroutine

### 3. Go Memory Model