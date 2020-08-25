package random

// For production use:
// Standard: Seeded with the current time in nanoseconds.
// Locked: Wrapped in a mutex to protect concurrent access.
//
// For tests:
// Dummy: Random, but deterministic output, same as Seeded(0).
// Mock: Mocked, needing to be configured before using.
// Seeded: Random, but deterministic output.
// Static: Always the same output.
