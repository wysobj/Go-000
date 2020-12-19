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
}()     // Anonymous
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
// 不推介，servePortA或者servePortB任一方退出了，另外的一方无法感知，导致出现部分服务可用
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
Go内存模型(Memory Model)主要说明多线程场景下，一个Goroutine的"写入(write)"结果能被另一个Goroutine的"读取(read)"正确读到需满足的条件
#### 3.1 可见性(Visibility)
一个线程写入操作的结果能否被另一个线程的读操作读到。造成"不可见"的原因主要包括：CPU的L1/L2 Cache多核心间不共享、编译或运行时的指令重排

#### 3.2 有序性
多个线程并发的操作指令是否具备Linear Serilization性质，也即线程的执行时序确定的情况下，指令的执行结果也是确定的，执行结果是确定的(Determined)。
> 单个线程内的操作都是有序的；从一个线程观察另一个线程，所有操作都不是有序的
```
var a, b int

func setup() {
	a = 1
	b = 2	
}

func main() {
	go setup()
	print(a)
	print(b)
}
```
上述程序的执行结果可能是"02"，原因就是setup的Goroutine内可能存在指令重排

#### 3.3 原子性(Atomicity)
一个操作的结果对外只体现"发生"和"未发生"，不会存在"部分发生"的情况

#### 3.4 Happens Before
Go Memory Model定义了"Happens Before"的概念，用以说明Goroutine中可见性需满足的条件
- 如果*w* does not happen after *r*且中间没有其他的写入，则*r*允许(allowed)看到*w*的写入
- 如果*w* happen before *r*且中间没其他写入，则*r*能确保(guaranteed)看到*w*的写入

### 4. Go Concurrency
#### 4.1 检测Data Race
```
go build -race
go test -race
```
#### 4.2 Atomic
容易误认为是原子的操作:
- 自增/自减
- 32位设备上对64字节长度的变量赋值
- 接口类型变量赋值。interface中包含两个字段Type和Data，分别指向接口变量对应的变量类型和变量值本身，因此接口类型变量的赋值需将这两个字段都赋值，不是原子操作

Atomic解决的不止是原子性的问题，也解决了可见性问题。因此不能因为赋值操作是Single Machine Word操作就认为可以不使用Atomic。Atomic样例:
```
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

func subscribe(container *atomic.Value, wg *sync.WaitGroup) {
	fmt.Println("Subscriber start.")
	for {
		v := container.Load()
		if flag, _ := v.(bool); flag == true {
			break
		}
	}
	fmt.Printf("Detect exit flag at %v", time.Now())
	wg.Done()
}

func publish(container *atomic.Value, wg *sync.WaitGroup) {
	fmt.Println("Publisher start.")
	time.Sleep(1000)
	container.Store(true)
	wg.Done()
}

func main() {
	wg := sync.WaitGroup{}
	var container atomic.Value
	wg.Add(2)
	go subscribe(&container, &wg)
	go publish(&container, &wg)
	wg.Wait()
	fmt.Println("Finish, exit.")
}
```
此外，Atomic也多用在*Copy On Write*的场景，用来在Write操作执行OK后直接原子替换原对象。*Copy On Write*适用于**读多写少**的场景。
#### 4.3 Lock
Go的sync包里包含的锁类型有sync.Mutex和sync.RWMutex。Go锁的Happens Before语义如下:
- Mutex的前一次Unlock之前的操作Happens Before后一次获取Lock之后的操作
- RWMutext的前一次Unlock之前的操作Happens Before获取到RLock之后的操作
使用样例:
```
package main

import (
	"sync"
	"fmt"
)

var counter int

func worker (lock *sync.Mutex, wg *sync.WaitGroup) {
	lock.Lock()
	for i := 0; i < 100; i++ {
		counter++
	}
	lock.Unlock()
	wg.Done()
}

func main() {
	var lock sync.Mutex
	var wg sync.WaitGroup
	for i  := 0; i < 100; i++ {
		wg.Add(1)
		go worker(&lock, &wg)
	}
	wg.Wait()
	fmt.Printf("Final counter value: %d, expect 10000.\n", counter)
}
```
Mutex锁的几种类型:
- Barging: 释放锁时唤醒第一个等待的Goroutine，被唤醒的Goroutine不确保能获得锁，因为有可能争抢不过新来获取锁的Goroutine
- Handsoff: 释放锁前会确保下一个等待的Goroutine已经准备好获得锁的控制权
- Spin: 获取锁的Goroutine自旋等待
Handsoff方式相比Barging和Spin会更加公平，减轻锁饥饿的情况发生。Go 1.8采用Spin + Barging的方式，Go 1.9以后新增了锁饥饿检测，如果发现存在锁饥饿时会切换到Handsoff模式，使锁相比更加公平
```
import (
	"fmt"
	"sync"
	"time"
)

const workerNum int = 10
const totalRound int = 100000
var aquires [workerNum]int
var round int

func worker(id int, lock *sync.Mutex, wg *sync.WaitGroup) {
	for round < totalRound {
		lock.Lock()
		aquires[id]++
		round++
		time.Sleep(10)
		lock.Unlock()
	}
	wg.Done()
}

func main() {
	var lock sync.Mutex
	var wg sync.WaitGroup
	for i := 0; i < workerNum; i++ {
		wg.Add(1)
		go worker(i, &lock, &wg)
	}
	wg.Wait()
	fmt.Printf("Lock aquire counts: %v\n.", aquires)
}
```
#### 4.4 Once
多次并发调用，确保传入的函数只被执行一次，多用于初始化和配置。Once的Happen Before语义为:所有对once.Do(f)的调用Happens Before第一次执行f的结果返回
```
import (
	"sync"
	"fmt"
)

func sayGoodWords() {
	fmt.Println("Good words never say twice.")
}

func say(once *sync.Once, wg *sync.WaitGroup) {
	once.Do(sayGoodWords)
	wg.Done()
}

func main() {
	var once sync.Once
	var wg sync.WaitGroup
	wg.Add(2)
	go say(&once, &wg)
	go say(&once, &wg)
	wg.Wait()
}
```
#### 4.5 Channel
> Do not communicate by sharing memory, instead, share memory by communicating.

- Channel是Go语言内置的一种并发编程模型，功能上相当于提供了一个线程安全的消息队列
- Channel的类型主要有"Unbuffered Channel"和"Buffered Channel"
- Channel的Happens Before语义为: 对Channel对象的发送操作Happens Before该发送的内容从Channel中被接收
- 通过for range遍历Channel。如果Channel没有元素会阻塞，如果Channel被关闭则遍历结束
```
func generate(c chan int) {
	for i := 0; i < 5; i++ {
		c <- i
	}
	close(c)
}

func main() {
	c := make(chan int)
	go generate(c)
	for i := range c {
		fmt.Printf("Receive %d\n", i)
	}
}
```
- 通过select管理多个Channel，一般要用for select组合。select如果有default case则不会阻塞。
```
func nextSeq(c chan int, quit chan interface{}) {
	var i int
	for {
	select {
		case c <- i:
			i++
		case <- quit:
			fmt.Println("Seq generate quit.")
			return
		}
	}
}

func main() {
	c := make(chan int)
	quit := make(chan interface{})
	go nextSeq(c, quit)
	fmt.Println("First element: ", <-c)
	fmt.Println("Second element: ", <-c)
	quit <- nil
	close(c)
	close(quit)
}
```
#### 4.6 Errgroup
功能就是一个具备能返回第一个失败Goroutine err并取消所有其他Goroutine能力的WaitGroup
```
package main

import (
	"golang.org/x/sync/errgroup"
	"fmt"
	"errors"
)

func raiseError() error {
	fmt.Println("Raise error done.")
	return errors.New("meet error.")
}

func normalWork() error {
	fmt.Println("Normal work done.")
	return nil
}

func main() {
	var eg errgroup.Group
	eg.Go(raiseError)
	eg.Go(normalWork)
	err := eg.Wait()
	fmt.Printf("Final error: %v\n", err)
}
```
#### 4.7 Pool
- 适用于构造对象资源池的场景，减少GC开销。
- sync.Pool有两个方法，Get和Put，Get是从pool中获取一个可用对象，Put是往pool中放置一个可用对象。另外在pool结构体初始化时也支持传入New方法
  - 如果**未**设置New方法，当Get操作获取不到可用对象时返回nil
  - 如果设置了New方法，当Get操作获取不到可用对象时会通过New函数来创建新对象返回
sync.Pool使用样例:
```
import (
	"sync"
	"fmt"
)

type Resource struct {
	id int
}

func main() {
	var id int
	pool := sync.Pool{
		New: func() interface{} {
			id++
			return &Resource{
				id : id,
			}
		},
	}

	rsc := pool.Get()
	fmt.Println(rsc)
	pool.Put(rsc)
	rsc = pool.Get()
	fmt.Println(rsc)
	rsc2 := pool.Get()
	fmt.Println(rsc2)
}
```
#### 4.8 Context
Context主要用于级联取消、超时控制场景，可以通过**线程安全**且优雅的方式级联取消多个Goroutine线程的任务执行。Context最终形成的数据结构为一棵树，任一树节点的取消/超时会传递到该节点的所有子树节点Context。主要的方法如下:
```
func Background() Context
func TODO() Context

func WithCancel(parent Context) (ctx Context, cancel CancelFunc)
func WithDeadline(parent Context, deadline time.Time) (Context, CancelFunc)
func WithTimeout(parent Context, timeout time.Duration) (Context, CancelFunc)
func WithValue(parent Context, key, val interface{}) Context
```
- WithCancel和WithTimeout样例
```
package main

/*
 *  Context cancel和timeout样例demo，树状关系结构如下
 *  main
 *	|__procedure1		  -> Cancelled
 *	  |__procedure1Child
 *	|__procedure2		  -> Timeout
 * 
 *  执行结果:
 *   Do something in main.
 *   Do something in procedure2.
 *   Do something in procedure1.
 *   Do something in procedure1Child.
 *   Procedure1Child exit at: 2020-12-13 21:15:02.1156564 +0800 CST *   m=+0.001050001, err: context canceled.
 *   Procedure1 exit at: 2020-12-13 21:15:02.115377 +0800 CST m=+0.000770501, err: context canceled.
 *   Procedure2 exit at: 2020-12-13 21:15:07.1190359 +0800 CST m=+5.004429701, err: context deadline exceeded
 *   Main exit at 2020-12-13 21:15:07.1193277 +0800 CST m=+5.004721401, err: %!s(<nil>)
 */

import (
	"fmt"
	"context"
	"sync"
	"time"
)

func procedure2(ctx context.Context, wg *sync.WaitGroup) {
	fmt.Println("Do something in procedure2.")
	<-ctx.Done()
	fmt.Printf("Procedure2 exit at: %s, err: %s\n", time.Now(), ctx.Err())
	wg.Done()
}

func procedure1Child(ctx context.Context) {
	fmt.Println("Do something in procedure1Child.")
	<-ctx.Done()
	fmt.Printf("Procedure1Child exit at: %s, err: %s.\n", time.Now(), ctx.Err())
}

func procedure1(ctx context.Context, wg *sync.WaitGroup) {
	fmt.Println("Do something in procedure1.")
	go procedure1Child(ctx)
	<-ctx.Done()
	fmt.Printf("Procedure1 exit at: %s, err: %s.\n", time.Now(), ctx.Err())
	wg.Done()
}

func main() {
	var wg sync.WaitGroup
	ctx := context.Background()
	ctx1, cancel1 := context.WithCancel(ctx)
	ctx2, _ := context.WithTimeout(ctx, 5 * time.Second)
	wg.Add(2)
	go procedure1(ctx1, &wg)
	go procedure2(ctx2, &wg)
	fmt.Println("Do something in main.")
	cancel1()
	wg.Wait()
	fmt.Printf("Main exit at %s, err: %s\n", time.Now(), ctx.Err())
}
```
- WithValue样例
```
package main

import (
	"context"
	"fmt"
)

func main() {
	ctxRoot := context.WithValue(context.Background(), "ctx_root", "ctx_root_value")
	ctx1 := context.WithValue(ctxRoot, "ctx1_key", "ctx1_value")
	ctx2 := context.WithValue(ctxRoot, "ctx2_key", "ctx2_value")

	fmt.Printf("ctx1[ctx_root]: %v, ctx1[ctx_1]: %v, ctx1[ctx_2]: %v\n", ctx1.Value("ctx_root"), ctx1.Value("ctx1_key"), ctx1.Value("ctx2_key"))
	fmt.Printf("ctx2[ctx_root]: %v, ctx2[ctx_1]: %v, ctx2[ctx_2]: %v\n", ctx2.Value("ctx_root"), ctx2.Value("ctx1_key"), ctx2.Value("ctx2_key"))
}
```