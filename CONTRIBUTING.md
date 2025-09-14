# Contributing to StreamV2

Thank you for your interest in contributing to StreamV2! 

## Current Development Status

**⚠️ StreamV2 is currently in active development with a stabilizing API.**

The design and API are still evolving as we refine the core functionality and establish best practices. We are not yet ready to accept external contributions while we work through:

- **API stabilization** - Core function signatures and interfaces
- **Design decisions** - Architecture patterns and performance optimizations  
- **Test suite completion** - Ensuring comprehensive coverage and fixing existing test failures
- **Documentation refinement** - Establishing clear usage patterns and examples
- **Performance optimization** - Auto-parallelization heuristics and executor backends

## When We'll Be Ready for Contributors

We expect to open for contributions once we have:

- ✅ **Stable API** - Core interfaces won't change significantly
- ✅ **Complete test suite** - All tests passing with high coverage
- ✅ **Performance benchmarks** - Established baseline performance metrics
- ✅ **Clear roadmap** - Defined feature priorities and development direction
- ✅ **Contribution guidelines** - Detailed process for submitting changes

This is expected to happen in **Q2 2024** (tentative).

## How You Can Help Now

While we're not accepting code contributions yet, you can still help by:

### 🧪 **Testing and Feedback**
- Try the [codelab tutorial](STREAMV2_CODELAB.md) 
- Test StreamV2 with your data processing use cases
- Report bugs and API pain points via [GitHub Issues](https://github.com/rosscartlidge/streamv2/issues)
- Share performance feedback on real workloads

### 📝 **Documentation**
- Suggest improvements to the [API documentation](docs/api.md)
- Report unclear or missing documentation
- Share examples of how you're using StreamV2

### 💡 **Feature Requests**
- Suggest stream operations that would be valuable
- Propose I/O format support (beyond CSV/JSON)
- Share performance optimization ideas

### 🐛 **Bug Reports**
When reporting bugs, please include:
- Go version and operating system
- Minimal code example that reproduces the issue
- Expected vs actual behavior
- Performance characteristics if relevant

## Staying Updated

- **Watch this repository** for updates on contribution status
- **Follow releases** for API stabilization milestones
- **Check issues** for development progress

## Future Contribution Areas

When we open for contributions, we'll be looking for help with:

### Core Functionality
- Stream operation implementations
- Performance optimizations
- Memory usage improvements
- Error handling enhancements

### I/O Operations
- Additional format support (Parquet, Avro, etc.)
- Database connectors (PostgreSQL, MySQL, etc.)
- Cloud storage integrations (S3, GCS, etc.)
- Streaming data sources (Kafka, Kinesis, etc.)

### Advanced Features
- Complex windowing operations
- Distributed processing capabilities
- GPU acceleration (CUDA integration)
- Custom aggregator frameworks

### Ecosystem Integration
- Database drivers
- Message queue connectors
- Observability and monitoring
- CLI tools and utilities

### Testing and Quality
- Performance benchmarks
- Stress testing frameworks
- Integration test suites
- Documentation examples

## Development Principles

When contributions do open, they should align with StreamV2's core principles:

### **Type Safety**
- Leverage Go's generics for compile-time safety
- Avoid `interface{}` where possible
- Clear, expressive function signatures

### **Performance**
- Lazy evaluation by default
- Memory efficiency for large datasets
- Auto-parallelization for appropriate workloads
- Minimal allocation in hot paths

### **Simplicity**
- Clean, composable API design
- Readable code with clear intent
- Comprehensive error handling
- Minimal external dependencies

### **Compatibility**
- Stable public API once established
- Backward compatibility within major versions
- Clear migration paths for breaking changes

## Questions?

If you have questions about:
- **Using StreamV2**: Check the [codelab](STREAMV2_CODELAB.md) and [API docs](docs/api.md)
- **Performance**: See the [performance guide](docs/performance.md)
- **Future contributions**: Open a [GitHub Discussion](https://github.com/rosscartlidge/streamv2/discussions)
- **Bug reports**: Use [GitHub Issues](https://github.com/rosscartlidge/streamv2/issues)

## Thank You

We appreciate your interest in StreamV2! While we're not ready for code contributions yet, your feedback, testing, and patience during this development phase are invaluable.

We're building StreamV2 to be the premier stream processing library for Go, and when we do open for contributions, we want to ensure a smooth, productive experience for everyone involved.

Stay tuned for updates! 🚀

---

**StreamV2 Team**  
*Last updated: January 2024*