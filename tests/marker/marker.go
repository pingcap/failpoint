// Copyright 2019 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package marker

import (
	"fmt"
	"math/rand"

	failpoint2 "github.com/pingcap/failpoint"
)

func markerBasic() {
	if ok, val := failpoint2.Eval("test6"); ok {
		fmt.Println("example", val)
	}
}

func markerAssign() {
	var _, f1, f2 = 10, func() {
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
	}, func() {
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
	}
	f1()
	f2()
}

func markerAssig2n() {
	_, f1, f2 := 10, func() {
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
	}, func() {
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
	}
	f1()
	f2()
}

func makerInGo() {
	go func() {
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
	}()
}

func makerInGo2() {
	go func(_ func()) {
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
	}(func() {
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
	})
}

func makerDefer() {
	defer func() {
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
	}()
}

func makerDefer2() {
	defer func(_ func()) {
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
	}(func() {
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
	})
}

func makerReturn() (func(), int) {
	return func() {
			if ok, val := failpoint2.Eval("test6"); ok {
				fmt.Println("example", val)
			}
		}, func() int {
			if ok, val := failpoint2.Eval("test6"); ok {
				fmt.Println("example", val)
			}
			return 1000
		}()
}

func markerIf1() {
	x := rand.Float32()
	if x > 0.5 {
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
	} else if x > 0.2 {
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
	} else {
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
	}
}

func markerIf2() {
	if a, b := func() {
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
	}, func() int { return rand.Intn(200) }(); b > 100 {
		a()
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
	}
}

func markerIf3() {
	if a, b := func() {
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
	}, func() int { return rand.Intn(200) }(); b > func() int {
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
		return rand.Intn(3000)
	}() && b < func() int {
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
		return rand.Intn(6000)
	}() {
		a()
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
	}
}

func markerSwitchCase() {
	switch x, y := rand.Intn(10), func() int { return rand.Intn(1000) }(); x - y + func() int {
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
		return rand.Intn(50)
	}() {
	case func() int {
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
		return rand.Intn(5)
	}(), func() int {
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
		return rand.Intn(8)
	}():
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
	default:
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
	}
}

func markerSwitchCase2() {
	switch x, y := rand.Intn(10), func() int {
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
		return rand.Intn(1000)
	}(); func(x, y int) int {
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
		return rand.Intn(50) + x + y
	}(x, y) {
	case func() int {
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
		return rand.Intn(5)
	}(), func() int {
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
		return rand.Intn(8)
	}():
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
	default:
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
		fn := func() {
			if ok, val := failpoint2.Eval("test6"); ok {
				fmt.Println("example", val)
			}
		}
		fn()
	}
}

func markerSelect() {
	select {
	case <-func() chan bool {
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
		return make(chan bool)
	}():
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}

	case <-func() chan bool {
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
		return make(chan bool)
	}():
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
	}
}

func markerLoop() {
	for i := func() int {
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
		return rand.Intn(100)
	}(); i < func() int {
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
		return rand.Intn(10000)
	}(); i += func() int {
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
		return rand.Intn(100)
	}() {
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
	}
}

func markerRange() {
	for x, y := range func() map[int]int {
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
		return map[int]int{}
	}() {
		if ok, val := failpoint2.Eval("test6"); ok {
			fmt.Println("example", val)
		}
		fn := func() {
			if ok, val := failpoint2.Eval("test6"); ok {
				fmt.Println("example", val, x, y)
			}
		}
		fn()
	}
}
