## **Week 2**

#### 1. 错误处理机制
- C : 通过函数返回值返回整数类型错误码表示具体错误
  - 缺点
    - 作为单返回值的语言，通过函数返回值返回错误限制只能通过参数列表的出参输出函数的处理数据，可读性和使用上存在不便
    - 不便于携带错误相关的上下文信息
- JAVA : 通过语言内置的异常机制(Checked Exception和Unchecked Exception)实现错误处理控制流
  - 优点
    - Checked Exception异常类型作为方法签名的一部分，能够很方便地知道方法会抛出的异常
    - 对于Checked Exception，在编译时强制约束调用者处理异常(捕捉或向上抛出)
  - 缺点
    - 每个Checked Exception都需要层层包含在方法签名里(如果不吞掉的话)，造成较大的编码负担
    - 异常机制提供了更灵活、非线性的控制流，但同时也引入了更高的复杂性，很容易错误地使用
        > My point isn’t that exceptions are bad. My point is that exceptions are too hard and I’m not 
smart enough to handle them.
- GO : Go中的错误就是一个实现了error interface的对象，一般情况通过函数返回值返回错误。当程序出现不可修复的错误时也支持Panic机制实现错误快速透传
  - 优点
    - Go支持多参数返回，因此通过返回值返回error不会带来不便
    - 直观简单，没有隐藏的控制流
    - 使用灵活，可以通过封装等手段来实现更丰富的功能
        > Error are values. 
  - 缺点
    - Go中的error机制更像一种约定而不是机制，因此对使用者的水平和熟悉程度有一定要求
    - 容易隐藏一些难以发现的Bug，例如错误地封装了Sentinel Error对象等

#### 2. Go中error处理手段
- Sentinel Error
  - 实现 : 通过在包中提供预定义的Error对象，包对外暴露该对象
  - 缺点 : 
    - 包对外暴露了具体类型，使得该Sentinel Error成为了包对外接口的一部分，形成了强耦合，无法保证调用者会用什么方式使用该对象，如果后续要修改该对象类型，极端情况甚至会影响调用者的功能。违反了"面向抽象而非具体"的编程思想
    - 不能携带更丰富的上下文 
    - 容易引入隐蔽的错误，例如返回路径上某个地方错误地封装了该对象
- Error Types
  - 实现 : 自定义实现Error方法的error类型，包对外暴露该类型
    - 缺点 :
      - 编码麻烦，判断是否为某种类型的错误需要使用类型断言
      - 和Sentinel一样，对外暴露了具体类型，引入调用者和被调用者之间的耦合
- Opaque Error
  - 实现 : 不对外暴露error的类型，也不提供预定义的error对象，而是通过对外暴露接口来让调用者检查是否是某种类型的错误
    - 优点 : 对外暴露的是抽象而非具体实现，对于判断错误类型的逻辑包提供者有完全的控制权，降低后续的维护难度

#### 3. 常见的错误编码样例
1. 主路径放在了缩进的代码中，可读性差
```
  foo, err := bar()
  if err == nil {
      // main procedure
  }

  // handle error
```
2. 不必要的返回值转换
```
  err := foo()
  if err != nil {
      return err
  }

  return nil
```
3. 传递闭包外的变量到闭包中，变量值在闭包外被修改
```
for i := 0; i < 10; i++ {
  go func() {
    print(i, "\n")
  }
}
```
4. 未对复杂的错误处理流程做较好的封装，影响代码整洁性和可读性。如使用io.Reader读取文件行数和使用bufio的Scanner来读取

#### 4. 提高Error的可定位性
1. fmt.Errorf的方式 : 不能带上调用栈的上下文信息，Sentinel Error不兼容
2. error返回路径上每层函数都打印错误日志 : 引入大量日志打印代码，且错误信息碎片化
3. 使用"github.com/pkg/errors"包 : 推荐
```
package main

import (
	"errors"
	"fmt"

	xerrors "github.com/pkg/errors"
)

var sentinelErr = errors.New("Sentinel Err!")

func level2() error {
	return xerrors.Wrapf(sentinelErr, "Raise at level2.")
}

func level1() error {
	return level2()
}

func main() {
	err := level1()
	fmt.Printf("main: %+v\n", err)
	fmt.Printf("cause: %v\n", xerrors.Cause(err))
}
```
