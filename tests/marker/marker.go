package marker

import (
	"context"
	"fmt"

	failpoint2 "github.com/pingcap/failpoint"
)

func markerBasic() {
	failpoint2.Marker("test", func(_ context.Context, arg *failpoint2.Arg) {
		fmt.Println("hello world", arg)
	})
}
