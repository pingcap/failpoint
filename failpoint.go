package failpoint

// Marker marks a fail point routine, which will be rewrite to a `if` statement
// and be triggered by fail point name specified `fpname`
func Marker(fpname string, fpblock interface{}) {}

// Break will generate a break statement in a loop, e.g:
// case1:
//   for i := 0; i < max; i++ {
//       failpoint.Marker("break-if-index-equal-2", func() {
//           if i == 2 {
//               failpoint.Break()
//           }
//       }
//   }
// failpoint.Break() => break
//
// case2:
//   outer:
//   for i := 0; i < max; i++ {
//       for j := 0; j < max / 2; j++ {
//           failpoint.Marker("break-if-index-i-equal-j", func() {
//               if i == j {
//                   failpoint.Break("outer")
//               }
//           }
//       }
//   }
// failpoint.Break("outer") => break outer
func Break(label ...string) {}

// Goto will generate a goto statement the same as `failpoint.Break()`
func Goto(label string) {}

// Continue will generate a continue statement the same as `failpoint.Break()`
func Continue(label ...string) {}

// Label will generate a label statement, e.g.
// case1:
//   failpoint.Label("outer")
//   for i := 0; i < max; i++ {
//       for j := 0; j < max / 2; j++ {
//           failpoint.Marker("break-if-index-i-equal-j", func() {
//               if i == j {
//                   failpoint.Break("outer")
//               }
//           }
//       }
//   }
// failpoint.Label("outer") => outer:
// failpoint.Break("outer") => break outer
func Label(label string) {}
