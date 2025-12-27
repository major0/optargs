# OptArgs Performance Report

Generated: Fri Dec 26 04:25:10 PM PST 2025
Go Version: go version go1.23.4 linux/amd64
System: Linux mf-desktop 6.8.0-88-generic #89-Ubuntu SMP PREEMPT_DYNAMIC Sat Oct 11 01:02:46 UTC 2025 x86_64 x86_64 x86_64 GNU/Linux

## Performance Baselines

Baseline file exists with 8 test cases

### Latest Baselines

- GetOpt_SimpleShortOptions: 8333 ns/op, 88 allocs/op, 2776 bytes/op
- GetOpt_CompactedOptions: 7772 ns/op, 86 allocs/op, 2712 bytes/op
- GetOpt_WithArguments: 7067 ns/op, 71 allocs/op, 2544 bytes/op
- GetOptLong_LongOptionsOnly: 6826 ns/op, 35 allocs/op, 2183 bytes/op
- GetOptLong_EqualsForm: 7239 ns/op, 35 allocs/op, 2182 bytes/op
- GetOptLong_MixedShortLong: 9861 ns/op, 73 allocs/op, 2805 bytes/op
- Iterator_FullConsumption: 11411 ns/op, 129 allocs/op, 3376 bytes/op
- Iterator_PartialConsumption: 6860 ns/op, 68 allocs/op, 2480 bytes/op

## Test Results

