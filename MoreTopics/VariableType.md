### 关于Go变量

#### 1. 值类型和引用类型
关于Go的值类型和引用类型，网上的总结如下:
> 值类型分别有：int系列、float系列、bool、string、数组和结构体
引用类型有：指针、slice切片、管道channel、接口interface、map、函数等
值类型的特点是：变量直接存储值，内存通常在栈中分配
引用类型的特点是：变量存储的是一个地址，这个地址对应的空间里才是真正存储的值，内存通常在堆中分配

如果使用过C和Java，在使用Go语言变量时不可避免地会将Go语言的变量使用规则同C或Java做类比，这里暂且把变量的使用规则称为"变量的编程模型"(想不到其他描述)。这里对Go的变量编程模型做一个补充:
| 差异项 | 值类型 | 引用类型 |
| --- | --- | --- |
| 变量语义 | 变量值为内存中保存的具体值 | 变量值为内存的地址 |
| 零值 | 对应具体类型的零值，例如int为0，bool为false | 零值为nil |
| 变量的定义 | 默认零值或通过字面量(Literal)方式赋值 | slice、map支持字面量方式赋值，通常使用内置的new、make语句 |
| 内存分配 | 无逃逸情况下，直接在栈上分配 | 在堆上分配内存，变量指向堆上的内存地址 |

#### 2. 局部变量的内存在哪分配
##### 2.1 默认情况下的内存分配行为
根据变量类型不同会在不同地方分配内存。默认情况下，值类型变量会直接在栈上分配内存，引用类型变量则在堆上分配内存。
```
package main

func main() {
	var anchor1 int
	var slice1 []int = []int{ 1, 2 } // composite literal
	var slice2 []int = make([]int, 2)
	var anchor2 int
	var map1 map[int]string = map[int]string{ 1:"1", 2:"2" }
	var map2 map[int]string = make(map[int]string, 2)
	var anchor3 int

	print("output variant on different devices:\n")
	print("anchor1: ", &anchor1, "\n")
	print("slice1: ", slice1, "\n")
	print("slice2: ", slice2, "\n")
	print("anchor2: ", &anchor2, "\n")
	print("map1: ", map1, "\n")
	print("map2: ", map2, "\n")
	print("anchor3: ", &anchor3, "\n")
}

/*
output variant on different devices:
anchor1: 0xc000042528
slice1: [2/2]0xc000042540
slice2: [2/2]0xc000042530
anchor2: 0xc000042520
map1: 0xc000042580
map2: 0xc000042550
anchor3: 0xc000042518
*/
```
从程序运行结果可以看出，变量anchor*是int类型，在栈上分配，anchor1、anchor2和anchor3的地址空间是连续的。但是slice1/slice2、map1/map2的地址空间和anchor1/anchor2/anchor3的地址空间是不连续的，因此可以推断，slice和map等引用类型变量的定义，无论是通过复合字面量(Composite Literal)还是make语句，其内存都是堆上分配。

##### 2.2 变量逃逸
变量逃逸分析是Go语言内置支持的一种机制。先看下面这个例子:
```
package main

type escapeStruct struct {
}

func escapeFunc() *escapeStruct {
	var es escapeStruct
	print("Inner func: ", &es, "\n")
	return &es
}

func main() {
	es := escapeFunc()
	print("Main: ", es, "\n")
}

/*
Inner func: 0xc000042750
Main: 0xc000042750
*/
```
按照C语言的编程模型，escapeFunc的局部变量es的内存会在函数escapeFunc执行结束后随着函数栈帧的销毁而释放，因此在main中访问返回的指针es会访问到非法内存地址。而在Go语言中这样编写程序是合理的，因为Go会对局部变量做逃逸分析，当检测到局部变量的作用域逃逸出了函数的作用域，则会在堆上分配该局部变量的内存。使用如下方式运行程序观察变量逃逸分析的结果，可以看到变量es被分配到了堆上:
```
go run --gcflags -m variable_escap_demo.go
# command-line-arguments
./variable_escap_demo.go:6:6: can inline escapeFunc
./variable_escap_demo.go:12:6: can inline main
./variable_escap_demo.go:13:18: inlining call to escapeFunc
./variable_escap_demo.go:7:6: moved to heap: es
Inner func: 0xc000042747
Main: 0xc000042748
```

##### 2.3 一个小插曲
在编写试验代码时，遇到一个奇怪的情况，对下面程序进行变量逃逸分析:
```
package main

import "fmt"

func main() {
        var i int
        fmt.Printf("%p\n", &i)
}
```
执行结果如下:
```
# command-line-arguments
./variable_escap_demo.go:7:12: inlining call to fmt.Printf
./variable_escap_demo.go:6:6: moved to heap: i
./variable_escap_demo.go:7:21: &i escapes to heap
./variable_escap_demo.go:7:12: main []interface {} literal does not escape
./variable_escap_demo.go:7:12: io.Writer(os.Stdout) escapes to heap
<autogenerated>:1: (*File).close .this does not escape
0xc0000a2010
```
可以看到局部变量i也被分析到逃逸，被分配在了堆上，而实际上变量i的作用域并没有超出函数main的作于域范围。后来查找发现这是一个目前还未关闭的[issue](https://github.com/golang/go/issues/8618)。


#### 3. 动态申请内存: new和make
- malloc: malloc是C语言申请内存使用的函数，其本身不是系统调用。该函数的作用是分配一块指定大小的内存，并将分配的内存块地址返回给调用程序。malloc不会对申请到的内存块中的数据做清零或初始化操作
- new: Go语言的new也是在堆上分配内存，会将分配到的内存清零
- make: make是Go语言内置的语句，专用于slice、map和channel这几种类型变量的内存分配，make不仅会对申请到的内存初始化清零，还会做对应的初始化操作，初始化slice、map和channel数据结构需要使用到的元数据(metadata)