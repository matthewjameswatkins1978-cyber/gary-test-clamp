# GARY Clamp Test Fixture

**Five simple assertions. One deliberately broken function.**

This is a small Go fixture for testing whether a coding worker can identify a straightforward regression, change the right code, and prove the fix with tests. It is **not** a finished general-purpose clamp library.

## The challenge

`Clamp(value, min, max)` should return the given value when it falls inside the inclusive range, the lower bound when it falls below, and the upper bound when it falls above.

The current implementation in [`clamp.go`](clamp.go) intentionally returns the input unchanged. The tests in [`clamp_test.go`](clamp_test.go) cover five cases:

| Input | Expected |
| --- | --- |
| `Clamp(-5, 0, 10)` | `0` |
| `Clamp(15, 0, 10)` | `10` |
| `Clamp(5, 0, 10)` | `5` |
| `Clamp(0, 0, 10)` | `0` |
| `Clamp(10, 0, 10)` | `10` |

## Run the fixture

Requires Go **1.24** or later.

```sh
go test ./...
```

**Expected starting result: failing tests.** The below-minimum and above-maximum cases should fail until someone fixes the function. That red test state is the fixture's point, not an accidental broken release.

An agent's task is to make the smallest correct change, run the suite, and report the observed outcome. If using this repository as a repeatable test, reset the deliberate bug before the next run.

There are no external Go dependencies and no licence file.
