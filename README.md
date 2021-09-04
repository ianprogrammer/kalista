# WIP: Kalista

The objective of this  `toy` project is to make easy to run contract tests 
where you don't need to write the code to run your tests you just need to define an config file with the definition of your contract test

# Example 

See contract definition example on folder [contracts](https://github.com/ianprogrammer/kalista/tree/main/contracts)

```go
package main

import (
	"github.com/ianprogrammer/kalista/internal/kalista"
)

func main() {

	k := kalista.NewKalista("./contracts")
	k.StartContracTest()
}


```
Kalista will validate all your contract definition and output a result like bellow

![image](https://user-images.githubusercontent.com/9167871/132080466-1ca62743-ed35-4caf-81c9-e940ea823a63.png)
