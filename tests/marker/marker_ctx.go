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
	"context"
	"fmt"
	"math/rand"

	failpoint2 "github.com/pingcap/failpoint"
)

var ctx = context.Background()

func markerBasicCtx() {
	if ok, val := failpoint2.Eval("test6", ctx); ok {
		fmt.Println("example", val)
	}
}

func markerAssignCtx() {
	var _, f1, f2 = 10, func() {
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
	}, func() {
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
	}
	f1()
	f2()
}

func markerAssig2nCtx() {
	_, f1, f2 := 10, func() {
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
	}, func() {
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
	}
	f1()
	f2()
}

func makerInGoCtx() {
	go func() {
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
	}()
}

func makerInGo2Ctx() {
	go func(_ func()) {
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
	}(func() {
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
	})
}

func makerDeferCtx() {
	defer func() {
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
	}()
}

func makerDefer2Ctx() {
	defer func(_ func()) {
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
	}(func() {
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
	})
}

func makerReturnCtx() (func(), int) {
	return func() {
			if ok, val := failpoint2.Eval("test6", ctx); ok {
				fmt.Println("example", val)
			}
		}, func() int {
			if ok, val := failpoint2.Eval("test6", ctx); ok {
				fmt.Println("example", val)
			}
			return 1000
		}()
}

func markerIf1Ctx() {
	x := rand.Float32()
	if x > 0.5 {
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
	} else if x > 0.2 {
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
	} else {
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
	}
}

func markerIf2Ctx() {
	if a, b := func() {
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
	}, func() int { return rand.Intn(200) }(); b > 100 {
		a()
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
	}
}

func markerIf3Ctx() {
	if a, b := func() {
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
	}, func() int { return rand.Intn(200) }(); b > func() int {
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
		return rand.Intn(3000)
	}() && b < func() int {
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
		return rand.Intn(6000)
	}() {
		a()
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
	}
}

func markerSwitchCaseCtx() {
	switch x, y := rand.Intn(10), func() int { return rand.Intn(1000) }(); x - y + func() int {
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
		return rand.Intn(50)
	}() {
	case func() int {
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
		return rand.Intn(5)
	}(), func() int {
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
		return rand.Intn(8)
	}():
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
	default:
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
	}
}

func markerSwitchCase2Ctx() {
	switch x, y := rand.Intn(10), func() int {
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
		return rand.Intn(1000)
	}(); func(x, y int) int {
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
		return rand.Intn(50) + x + y
	}(x, y) {
	case func() int {
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
		return rand.Intn(5)
	}(), func() int {
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
		return rand.Intn(8)
	}():
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
	default:
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
		fn := func() {
			if ok, val := failpoint2.Eval("test6", ctx); ok {
				fmt.Println("example", val)
			}
		}
		fn()
	}
}

func markerSelectCtx() {
	select {
	case <-func() chan bool {
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
		return make(chan bool)
	}():
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}

	case <-func() chan bool {
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
		return make(chan bool)
	}():
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
	}
}

func markerLoopCtx() {
	for i := func() int {
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
		return rand.Intn(100)
	}(); i < func() int {
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
		return rand.Intn(10000)
	}(); i += func() int {
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
		return rand.Intn(100)
	}() {
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
	}
}

func markerRangeCtx() {
	for x, y := range func() map[int]int {
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
		return map[int]int{}
	}() {
		if ok, val := failpoint2.Eval("test6", ctx); ok {
			fmt.Println("example", val)
		}
		fn := func() {
			if ok, val := failpoint2.Eval("test6", ctx); ok {
				fmt.Println("example", val, x, y)
			}
		}
		fn()
	}
}
