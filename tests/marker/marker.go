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

func markerBasic() {
	failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
		fmt.Println("example", arg)
	})
}

func markerAssign() {
	var _, f1, f2 = 10, func() {
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
	}, func() {
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
	}
	f1()
	f2()
}

func markerAssig2n() {
	_, f1, f2 := 10, func() {
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
	}, func() {
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
	}
	f1()
	f2()
}

func makerInGo() {
	go func() {
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
	}()
}

func makerInGo2() {
	go func(_ func()) {
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
	}(func() {
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
	})
}

func makerDefer() {
	defer func() {
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
	}()
}

func makerDefer2() {
	defer func(_ func()) {
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
	}(func() {
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
	})
}

func makerReturn() (func(), int) {
	return func() {
			failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
				fmt.Println("example", arg)
			})
		}, func() int {
			failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
				fmt.Println("example", arg)
			})
			return 1000
		}()
}

func markerIf1() {
	x := rand.Float32()
	if x > 0.5 {
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
	} else if x > 0.2 {
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
	} else {
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
	}
}

func markerIf2() {
	if a, b := func() {
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
	}, func() int { return rand.Intn(200) }(); b > 100 {
		a()
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
	}
}

func markerIf3() {
	if a, b := func() {
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
	}, func() int { return rand.Intn(200) }(); b > func() int {
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
		return rand.Intn(3000)
	}() && b < func() int {
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
		return rand.Intn(6000)
	}() {
		a()
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
	}
}

func markerSwitchCase() {
	switch x, y := rand.Intn(10), func() int { return rand.Intn(1000) }(); x - y + func() int {
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
		return rand.Intn(50)
	}() {
	case func() int {
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
		return rand.Intn(5)
	}(), func() int {
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
		return rand.Intn(8)
	}():
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
	default:
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
	}
}

func markerSwitchCase2() {
	switch x, y := rand.Intn(10), func() int {
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
		return rand.Intn(1000)
	}(); func(x, y int) int {
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
		return rand.Intn(50) + x + y
	}(x, y) {
	case func() int {
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
		return rand.Intn(5)
	}(), func() int {
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
		return rand.Intn(8)
	}():
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
	default:
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
		fn := func() {
			failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
				fmt.Println("example", arg)
			})
		}
		fn()
	}
}

func markerSelect() {
	select {
	case <-func() chan bool {
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
		return make(chan bool)
	}():
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})

	case <-func() chan bool {
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
		return make(chan bool)
	}():
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
	}
}

func markerLoop() {
	for i := func() int {
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
		return rand.Intn(100)
	}(); i < func() int {
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
		return rand.Intn(10000)
	}(); i += func() int {
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
		return rand.Intn(100)
	}() {
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
	}
}

func markerRange() {
	for x, y := range func() map[int]int {
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
		return map[int]int{}
	}() {
		failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
			fmt.Println("example", arg)
		})
		fn := func() {
			failpoint2.Marker("test6", func(_ context.Context, arg *failpoint2.Arg) {
				fmt.Println("example", arg, x, y)
			})
		}
		fn()
	}
}
