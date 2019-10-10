# io

this is a non-blocking multiple reader function

use this like go default MultiReader

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
