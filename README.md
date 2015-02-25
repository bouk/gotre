# gotre: Golang Tail Recursion Eliminator

Gotre takes a `.go` files and rewrites any functions that do tail recursion to prevent allocating unneeded stackframes. The go compiler currently doesn't do this automatically.

## Example

```go
// cat fib.go
package fib

func fibRecurse(a, b, n int) int {
  if n <= 0 {
    return a
  }
  return fibRecurse(b, a+b, n-1)
}

func Fib(n int) int {
  return fibRecurse(0, 1, n)
}

// ./gotre test/fib.go
package fib

func fibRecurse(a, b, n int) int {
__tre__:
  if n <= 0 {
    return a
  }
  a, b, n = b, a+b, n-1
  goto __tre__
}

func Fib(n int) int {
  return fibRecurse(0, 1, n)
}

```
