# Hash-Based Distribution System

## Purpose
This code implements a hash-based distribution mechanism to assign request IDs to backend groups consistently and uniformly. It supports diverse input formats, ensures stickiness, and provides fair distribution based on specified ratios.

## Core Features

### Flexible Input Handling
The system is designed to handle any type of input. Current examples include:
- Short integers (e.g., `123`)
- Long integers (e.g., `987654321`)
- Tool-prefixed IDs (e.g., `Tool-001`)
- Alphanumeric strings (e.g., `ID-abCdEf01`)

Other ID formats can be easily accommodated without modification to the core logic.

### Deterministic
- Ensures the same input always maps to the same group for a given distribution.

### Uniform Distribution
- Distributes inputs across groups based on predefined ratios (e.g., `50-50`, `70-30`).
- Achieves as close to uniformity as possible using the MD5 hash.

## Tests

### 1. `Test_SimpleDistribution`
- **Objective**: Validate the distribution logic for basic ratios.
- **Example Ratios**: `50-50`, `70-30`
- **Results**:

```vbnet
===== Testing 50-50 Distribution =====
Group 1: 4978 (49.78%)
Group 2: 5022 (50.22%)
Is there any ID distributed to more than 1 group? false
```

### 2. `Test_HashWithVariations`
- **Objective**: Test robustness with different input formats.
- **Input Examples**:
  - Short integers (e.g., `98`)
  - Long integers (e.g., `123456789012`)
  - Tool-prefixed strings (e.g., `Tool-123`)
  - Random alphanumeric strings (e.g., `ID-abCdEf01`)
- **Results**:

```yaml
===== Testing 50-50 Distribution with ID Type: short_int =====
Group 1: 4975 (49.75%)
Group 2: 5025 (50.25%)
```

```yaml
===== Testing 50-50 Distribution with ID Type: string =====
Group 1: 4980 (49.80%)
Group 2: 5020 (50.20%)
```

### 3. `Test_BenchmarkTiming`
- **Objective**: Measure performance of the distribution mechanism.
- **Results**:

```sql
Benchmark for ID Type short_int: Total time = 234.523ms, Average time per computation = 234ns
Benchmark for ID Type long_int: Total time = 230.275ms, Average time per computation = 230ns
Benchmark for ID Type tool: Total time = 234.947ms, Average time per computation = 234ns
Benchmark for ID Type string: Total time = 230.522ms, Average time per computation = 230ns
```

## Extensibility
- New ID formats can be easily integrated into the system by modifying or extending the `Test_GenerateRandomIDs` function.
- The hash-based logic remains agnostic to input type, ensuring future-proofing for diverse use cases.

## Enhancement Suggestions

### Use xxHash Instead of MD5
While the current implementation uses MD5, xxHash can be a better alternative for high-throughput systems. Here’s why:
- **Performance**: xxHash is significantly faster than MD5 and SHA256, making it ideal for high-performance applications.
- **Low Overhead**: Consumes less CPU and memory.
- **Uniformity**: Provides excellent distribution quality.

To replace MD5 with xxHash, update the hashing logic with a Go library like [cespare/xxhash](https://github.com/cespare/xxhash).
