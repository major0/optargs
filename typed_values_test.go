package optargs

// Compile-time interface satisfaction checks.
var (
	_ TypedValue = (*stringValue)(nil)
	_ TypedValue = (*boolValue)(nil)
	_ TypedValue = (*scalarValue[int])(nil)
	_ TypedValue = (*durationValue)(nil)
	_ BoolValuer = (*boolValue)(nil)
)
