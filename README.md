# io

This is a non-blocking multiple reader function

Use this like go default MultiReader

In order to fix go MultiReader, if Reader[1] has no data outputting, 
it will wait forever,instead of executing Reader[2].

If the data of Reader[2] comes at this time, sometimes it will cause some problems.

Please judge which MulitiReader to use according to your own needs.

## tips:
### If the read content is larger than the cache, the content will be cut off. 
```go
buff  := make([]byte,3)

b1:="1234"
b2:="abcdef"
reader := MultiReader(strings.NewReader(b1),strings.NewReader(b2))
```
if use this multi reader this perhaps:
>
  123 
  abc 
  4 
  def 

If you want to solve this problem, give him a bigger cache.
